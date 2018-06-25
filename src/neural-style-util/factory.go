package NSUtil

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consul "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	ht "github.com/go-kit/kit/transport/http"
)

// GetIPv4Host get the activate ipv4 host address
func GetIPv4Host() (string, error) {
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}
	return "", nil
}

// RegisterSDService register a general service discovery for a path and method
func RegisterSDService(ctx context.Context, r *mux.Router, client consul.Client, logger log.Logger, name, tag, method, path string,
	duration time.Duration, retryTimes int) *mux.Router {
	factory := registerFactory(ctx, method)
	tags := []string{name, tag}
	instancer := consul.NewInstancer(client, logger, name, tags, true)
	endpointer := sd.NewEndpointer(instancer, factory, logger)
	balancer := lb.NewRoundRobin(endpointer)
	retry := lb.Retry(retryTimes, duration, balancer)

	r.Methods(method).Path(path).Handler(ht.NewServer(
		retry,
		decodeServiceDiscoveryRequest,
		encodeServiceDiscoveryResponse,
	))

	return r
}

// GeneralResponse ...
type GeneralResponse struct {
	Body       []byte
	StatusCode int
}

func decodeServiceDiscoveryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return r, nil
}

func encodeServiceDiscoveryResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp, isOk := response.(GeneralResponse)
	if !isOk {
		return errors.New("Bad Request")
	}

	w.Header().Set("Status-Code", strconv.Itoa(resp.StatusCode))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	_, err := w.Write(resp.Body)
	return err
}

func registerFactory(_ context.Context, method string) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}

		tgt, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}

		var (
			enc ht.EncodeRequestFunc
			dec ht.DecodeResponseFunc
		)

		enc = func(_ context.Context, req *http.Request, request interface{}) error {
			inputReq, isOk := request.(http.Request)
			if !isOk {
				return errors.New("Bad Request")
			}

			// assign all the url except the host; the host is defined from instance string
			req.URL.Path = inputReq.URL.Path
			req.URL.RawQuery = inputReq.URL.RawQuery

			// assign all header
			req.Header = inputReq.Header

			// assign the body
			req.Body = inputReq.Body

			return nil
		}

		dec = func(_ context.Context, resp *http.Response) (interface{}, error) {
			var responseInfo GeneralResponse
			responseInfo.StatusCode = resp.StatusCode
			info, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				return nil, err
			}

			responseInfo.Body = info
			return responseInfo, nil
		}

		return ht.NewClient(method, tgt, enc, dec).Endpoint(), nil, nil
	}
}

// RegisterFactoryWithDecoderAndEncoder generate factory with the input encode and decoder
func RegisterFactoryWithDecoderAndEncoder(_ context.Context, method, path string,
	encoder ht.EncodeRequestFunc, decoder ht.DecodeResponseFunc) sd.Factory {

	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}

		tgt, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}

		tgt.Path = path

		return ht.NewClient(method, tgt, encoder, decoder).Endpoint(), nil, nil
	}
}
