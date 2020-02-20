/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ordermanager

import (
	context "context"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mattmoor/http01-solver/pkg/challenger"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/apis"
	logging "knative.dev/pkg/logging"
)

// Interface defines the interface for ordering new certificates.
type Interface interface {
	Order(ctx context.Context, domains []string, owner interface{}) (challenges []*apis.URL, cert *tls.Certificate, err error)
}

// OrderUpCallback is the signature of the function for notifying
// owners that their order is up, and that they should invoke Order
// again to pick it up.
type OrderUpCallback func(owner interface{})

const (
	Staging    = "https://acme-staging-v02.api.letsencrypt.org/directory"
	Production = autocert.DefaultACMEDirectory
)

var (
	// Endpoint is the ACME API to use, it defaults to Production, but
	// can be pointed at Staging or other compatible endpoints.
	Endpoint = Production
)

// New creates a new OrderManager.
func New(ctx context.Context, cb OrderUpCallback, chlr challenger.Interface) (Interface, error) {
	acctKey, err := ecdsa.GenerateKey(elliptic.P256(), cryptorand.Reader)
	if err != nil {
		return nil, err
	}
	a := &acme.Account{Contact: []string{}}
	client := &acme.Client{
		DirectoryURL: Endpoint,
		UserAgent:    "github.com/mattmoor/http01-solver",
		Key:          acctKey,
	}
	_, err = client.Register(ctx, a, autocert.AcceptTOS)
	if err != nil && err != acme.ErrAccountAlreadyExists {
		return nil, err
	}

	return &impl{
		Client:     client,
		Callback:   cb,
		Challenger: chlr,
		inflight:   make(map[key]ticket, 10),
	}, nil
}

// impl implements Interface.
type impl struct {
	sync.Mutex // guards access to inflight.

	Challenger challenger.Interface
	Client     *acme.Client
	Callback   OrderUpCallback

	inflight map[key]ticket
}

var _ Interface = (*impl)(nil)

// ticket is used to represent an unclaimed order that is working
// it's way through the system.
type ticket struct {
	uri   string
	owner interface{}
	err   error
}

// Order implements Interface
func (om *impl) Order(ctx context.Context, domains []string, owner interface{}) ([]*apis.URL, *tls.Certificate, error) {
	logger := logging.FromContext(ctx)
	t, ok := om.getTicket(domains)
	if !ok {
		// If there isn't an in-flight order, then initiate a new order.
		var err error
		t, err = om.initiateNewOrder(ctx, domains, owner)
		if err != nil {
			return nil, nil, err
		}
		// Fall through to return the challenges
	}
	if t.err != nil {
		om.cancelOrder(ctx, domains)
		logger.Infof("Cancelling order for %v due to error: %v", domains, t.err)
		return nil, nil, t.err
	}

	// See if the order specified by this ticket is ready.
	status, err := t.GetStatus(ctx, om.Client)
	if err != nil {
		// TODO(mattmoor): If the error isn't transient,
		// then we should clear the ticket here, otherwise
		// we are in an unrecoverable state.
		return nil, nil, err
	}
	switch status {
	case acme.StatusReady, acme.StatusValid:
		logger.Infof("Order is ready for %v", domains)
		// This removes the ticket, a subsequent Order will start
		// the process over.
		// TODO(mattmoor): Consider keeping completed orders around
		// until they have reached some level of staleness as a
		// precaution against the low rate limit of let's encrypt.
		cert, err := om.completeOrder(ctx, domains, t)
		return nil, cert, err

	case acme.StatusPending, acme.StatusProcessing, acme.StatusUnknown:
		logger.Infof("Order is pending for %v", domains)
		urls, err := t.ChallengeURLs(ctx, om.Client)
		return urls, nil, err

	case acme.StatusDeactivated, acme.StatusExpired, acme.StatusInvalid,
		acme.StatusRevoked:
		logger.Infof("Order is invalid for %v", domains)
		// This is a permanently bad state, we should flush the ticket
		// and return an error to the client which can retry as it sees
		// fit.
		om.cancelOrder(ctx, domains)
		if err1, err2 := t.GetError(ctx, om.Client); err2 != nil {
			// An error getting the error.
			logging.FromContext(ctx).Errorf("Error getting error: %v", err)
			return nil, nil, err2
		} else if err1 != nil {
			// The error returned by the CA leading to the above state.
			logging.FromContext(ctx).Errorf("Error from theh CA: %v", err)
			return nil, nil, err1
		}
		// Fallback on reporting the status.
		logging.FromContext(ctx).Errorf("Bad status for order: %s", status)
		return nil, nil, fmt.Errorf("Order resulted in status: %q", status)

	default:
		return nil, nil, fmt.Errorf("Unknown order status: %q", status)
	}
}

func (om *impl) getTicket(domains []string) (t ticket, found bool) {
	om.Lock()
	defer om.Unlock()

	key := asKey(domains)
	t, found = om.inflight[key]
	return
}

func (om *impl) initiateNewOrder(ctx context.Context, domains []string, owner interface{}) (ticket, error) {
	o, err := om.Client.AuthorizeOrder(ctx, acme.DomainIDs(domains...))
	if err != nil {
		logging.FromContext(ctx).Errorf("Error creating new order: %v", err)
		return ticket{}, err
	}

	eg := &errgroup.Group{}
	for _, zurl := range o.AuthzURLs {
		z, err := om.Client.GetAuthorization(ctx, zurl)
		if err != nil {
			return ticket{}, err
		}
		// Find the HTTP01 challenge (all we support)
		chal, err := getHTTP01(z.Challenges)
		if err != nil {
			return ticket{}, err
		}
		resp, err := om.Client.HTTP01ChallengeResponse(chal.Token)
		if err != nil {
			return ticket{}, err
		}
		path := om.Client.HTTP01ChallengePath(chal.Token)
		om.Challenger.RegisterChallenge(path, resp)

		eg.Go(func() error {
			defer om.Challenger.UnregisterChallenge(path)

			// TODO(mattmoor): Wait until we have successfully probed the
			// challenge ourselves before accepting to get positive hand-off
			// to the routing layer that things have been successfully plumbed.
			// something something wait.Until()

			time.Sleep(2 * time.Second)

			if _, err := om.Client.Accept(ctx, chal); err != nil {
				return err
			}
			if _, err := om.Client.WaitAuthorization(ctx, z.URI); err != nil {
				return err
			}
			return nil
		})
	}

	go func() {
		if err := eg.Wait(); err != nil {
			logging.FromContext(ctx).Errorf("Encountered an error waiting for challenges: %v.", err)
			om.setError(ctx, domains, err)
		} else if _, err := om.Client.WaitOrder(ctx, o.URI); err != nil {
			logging.FromContext(ctx).Errorf("Encountered an error waiting for order: %v.", err)
			om.setError(ctx, domains, err)
		} else {
			logging.FromContext(ctx).Infof("Order %q has completed without error.", o.URI)
		}

		// The order is ready for fulfillment (one way or another)!
		om.Callback(owner)
	}()

	logging.FromContext(ctx).Infof("Order %q has been initiated.", o.URI)
	t := ticket{
		uri:   o.URI,
		owner: owner,
	}
	func() {
		om.Lock()
		defer om.Unlock()
		om.inflight[asKey(domains)] = t
	}()
	return t, nil
}

func (om *impl) completeOrder(ctx context.Context, domains []string, t ticket) (*tls.Certificate, error) {
	cert, err := t.GetCertificate(ctx, om.Client, domains)
	if err != nil {
		return nil, err
	}

	om.Lock()
	defer om.Unlock()

	delete(om.inflight, asKey(domains))
	return cert, nil
}

func (om *impl) cancelOrder(ctx context.Context, domains []string) {
	om.Lock()
	defer om.Unlock()

	delete(om.inflight, asKey(domains))
}

func (om *impl) setError(ctx context.Context, domains []string, err error) {
	om.Lock()
	defer om.Unlock()

	t, ok := om.inflight[asKey(domains)]
	if !ok {
		return
	}
	t.err = err
	om.inflight[asKey(domains)] = t
}

func (t *ticket) GetStatus(ctx context.Context, client *acme.Client) (string, error) {
	o, err := client.GetOrder(ctx, t.uri)
	if err != nil {
		return "", err
	}
	return o.Status, nil
}

func (t *ticket) GetError(ctx context.Context, client *acme.Client) (error, error) {
	o, err := client.GetOrder(ctx, t.uri)
	if err != nil {
		return nil, err
	}
	if o.Error != nil {
		return o.Error, nil
	}
	return nil, nil
}

func (t *ticket) ChallengeURLs(ctx context.Context, client *acme.Client) ([]*apis.URL, error) {
	o, err := client.GetOrder(ctx, t.uri)
	if err != nil {
		return nil, err
	}
	urls := make([]*apis.URL, 0, len(o.AuthzURLs))

	// Satisfy all pending authorizations.
	for _, zurl := range o.AuthzURLs {
		z, err := client.GetAuthorization(ctx, zurl)
		if err != nil {
			return nil, err
		}
		// autocert skips authorizations that aren't pending,
		// but we include them to avoid churn.

		// Find the HTTP01 challenge (all we support)
		chal, err := getHTTP01(z.Challenges)
		if err != nil {
			return nil, err
		}

		urls = append(urls, &apis.URL{
			Scheme: "http",
			Host:   z.Identifier.Value,
			Path:   client.HTTP01ChallengePath(chal.Token),
		})
	}
	return urls, nil
}

func (t *ticket) GetCertificate(ctx context.Context, client *acme.Client, domains []string) (*tls.Certificate, error) {
	order, err := client.GetOrder(ctx, t.uri)
	if err != nil {
		return nil, err
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), cryptorand.Reader)
	if err != nil {
		return nil, err
	}
	req := &x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: domains[0]},
		DNSNames: domains,
	}
	csr, err := x509.CreateCertificateRequest(cryptorand.Reader, req, key)
	if err != nil {
		return nil, err
	}
	der, _, err := client.CreateOrderCert(ctx, order.FinalizeURL, csr, true)
	if err != nil {
		return nil, err
	}
	x509Cert, err := x509.ParseCertificates(flattenBytes(der))
	if err != nil || len(x509Cert) == 0 {
		return nil, err
	}
	return &tls.Certificate{
		Certificate: der,
		PrivateKey:  key,
		Leaf:        x509Cert[0],
	}, nil
}

var ErrHTTP01Unavailable = errors.New("The CA didn't list HTTP01 as a viable certificate challenge.")

func getHTTP01(challs []*acme.Challenge) (*acme.Challenge, error) {
	for _, c := range challs {
		if c.Type == "http-01" {
			return c, nil
		}
	}
	return nil, ErrHTTP01Unavailable
}

type key string

func asKey(domains []string) key {
	return key(strings.Join(sets.NewString(domains...).List(), ","))
}

// From acme/autocert
func flattenBytes(der [][]byte) []byte {
	var n int
	for _, b := range der {
		n += len(b)
	}
	pub := make([]byte, n)
	n = 0
	for _, b := range der {
		n += copy(pub[n:], b)
	}
	return pub
}
