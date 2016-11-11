package jsonrpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/prometheus/client_golang/prometheus"
)

type responseError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    string `json:"data,omitempty"`
}

type requestData struct {
	Id      string          `json:"id"`
	Version string          `json:"jsonrpc"`
	Params  json.RawMessage `json:"params"`
	Method  string          `json:"method"`
}

type responseData struct {
	Id      string         `json:"id,omitempty"`
	Version string         `json:"jsonrpc"`
	Result  interface{}    `json:"result,omitempty"`
	Error   *responseError `json:"error,omitempty"`
}

type ParametersInterface interface {
	Validate() error
}

type MethodInterface interface {
	Params() ParametersInterface
	Action(request *http.Request, params ParametersInterface) (response interface{}, err error)
}

type Service struct {
	methods        map[string]MethodInterface
	accept         []string
	callSummary    *prometheus.SummaryVec
	requestSummary *prometheus.SummaryVec
}

func NewService() *Service {
	calls := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "rpc_call_duration_microseconds",
			Help: "The duration of rpc method calls to the service",
		},
		[]string{"method", "error_code"},
	)
	requests := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "rpc_request_duration_microseconds",
			Help: "The duration of rpc requests to the service",
		},
		[]string{"error_code"},
	)
	prometheus.Register(calls)
	prometheus.Register(requests)
	service := &Service{
		methods:        map[string]MethodInterface{},
		accept:         []string{"application/json", "text/json"},
		callSummary:    calls,
		requestSummary: requests,
	}
	return service
}

func (service *Service) RegisterMethod(name string, method MethodInterface) (err error) {
	_, exists := service.methods[name]
	if exists {
		return fmt.Errorf("rpc: method name `%s` has already been registered on this service", name)
	}
	service.methods[name] = method
	return
}

func (service *Service) recordCall(method string, start time.Time, code int) {
	now := time.Now()
	duration := now.Sub(start)
	cs := fmt.Sprintf("%d", code)
	df := float64(duration) / float64(time.Microsecond)
	service.callSummary.WithLabelValues(method, cs).Observe(df)
}

func (service *Service) handleCall(request requestData, response *responseData, rawRequest *http.Request, wg *sync.WaitGroup, start time.Time) {
	var errCode int
	defer func() {
		wg.Done()
		service.recordCall(request.Method, start, errCode)
	}()
	errContext, errID := raven.CapturePanic(func() {
		method, ok := service.methods[request.Method]

		if !ok {
			errCode = -32601
			response.Error = &responseError{
				Code:    -32601,
				Message: fmt.Sprintf("rpc: Method name `%s` does not exist", request.Method),
			}
			return
		}

		params := method.Params()
		err := json.Unmarshal(request.Params, params)

		if err != nil {
			errCode = -32602
			response.Error = &responseError{
				Code:    -32602,
				Message: err.Error(),
			}
			return
		}

		err = params.Validate()

		if err != nil {
			errCode = -32602
			response.Error = &responseError{
				Code:    -32602,
				Message: err.Error(),
			}
			return
		}

		result, err := method.Action(rawRequest, params)

		if err != nil {
			errCode = -32603
			response.Error = &responseError{
				Code:    -32603,
				Message: err.Error(),
			}
			return
		}

		response.Result = result
		end := time.Now()
		fmt.Printf("method %s responded in %s\n", request.Method, end.Sub(start))
		return
	}, map[string]string{"method": request.Method})
	if errID != "" {
		errCode = -32700
		response.Error = &responseError{
			Code: -32700,
			Message: fmt.Sprintf(
				"rpc: panic occured and reported to sentry: %s", errID,
			),
		}
		switch errVal := errContext.(type) {
		case nil:
			break
		case error:
			response.Error.Data = errVal.Error()
		default:
			response.Error.Data = fmt.Sprint(errVal)
		}
	}
}

func (service *Service) recordRequest(start time.Time, code int) {
	now := time.Now()
	duration := now.Sub(start)
	cs := fmt.Sprintf("%d", code)
	df := float64(duration) / float64(time.Microsecond)
	service.requestSummary.WithLabelValues(cs).Observe(df)
}

func (service *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var errCode int
	defer func() {
		service.recordRequest(start, errCode)
	}()
	w.Header().Set("Content-Type", "application/json")
	errContext, errID := raven.CapturePanic(func() {

		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			errCode = -32700
			errR := responseData{
				Version: "2.0",
				Error: &responseError{
					Code:    -32700,
					Message: "rpc: invalid HTTP method",
				},
			}
			data, _ := json.Marshal(errR)
			w.Write(data)
			return
		}

		requests := []requestData{}
		responses := []*responseData{}

		data, err := ioutil.ReadAll(r.Body)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errCode = -32700
			errR := responseData{
				Version: "2.0",
				Error: &responseError{
					Code:    -32700,
					Message: err.Error(),
				},
			}
			data, _ = json.Marshal(errR)
			w.Write(data)
			return
		}

		err = json.Unmarshal(data, &requests)

		var single bool

		if err != nil {
			singleRequest := requestData{}
			sErr := json.Unmarshal(data, &singleRequest)
			if sErr != nil {
				w.WriteHeader(http.StatusBadRequest)
				errCode = -32700
				errR := responseData{
					Version: "2.0",
					Error: &responseError{
						Code:    -32700,
						Message: err.Error(),
					},
				}
				data, _ = json.Marshal(errR)
				w.Write(data)
				return
			}
			single = true
			requests = append(requests, singleRequest)
		}

		var wg sync.WaitGroup

		noRequests := len(requests)

		for _, request := range requests {
			response := new(responseData)
			response.Id = request.Id
			response.Version = "2.0"
			responses = append(responses, response)
			wg.Add(1)
			if noRequests == 1 {
				service.handleCall(request, response, r, &wg, start)
			} else {
				go service.handleCall(request, response, r, &wg, start)
			}
		}

		wg.Wait()

		if single && len(responses) == 1 {
			data, err = json.Marshal(responses[0])
		} else {
			data, err = json.Marshal(responses)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errCode = -32603
			errR := responseData{
				Version: "2.0",
				Error: &responseError{
					Code:    -32603,
					Message: err.Error(),
				},
			}
			data, _ = json.Marshal(errR)
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		end := time.Now()
		fmt.Printf(
			"request from %s (%s) processed in %s\n",
			r.UserAgent(), r.RemoteAddr, end.Sub(start),
		)
	}, nil)
	if errID != "" {
		w.WriteHeader(http.StatusInternalServerError)
		errCode = -32700
		errR := responseData{
			Version: "2.0",
			Error: &responseError{
				Code: -32700,
				Message: fmt.Sprintf(
					"rpc: panic occured and reported to sentry: %s", errID,
				),
			},
		}
		switch errVal := errContext.(type) {
		case nil:
			break
		case error:
			errR.Error.Data = errVal.Error()
		default:
			errR.Error.Data = fmt.Sprint(errVal)
		}
		data, _ := json.Marshal(errR)
		w.Write(data)
	}
}
