package mock

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func DummyAPI(id string) *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// TODO: eventually add in origins and pools if operator ever needs them.
		out := fmt.Sprintf(`{
			"data": {
				"loadBalancer": {
				  "id": "%s",
				  "name": "a-very-nice-lb",
				  "ports": {
					"edges": [
					  {
						"node": {
						  "id": "loadprt-2fiP_C_gnAORx_oDbNJAf"
						}
					  },
					  {
						"node": {
						  "id": "loadprt-Ox62077uY1igHFU1MIvyl"
						}
					  }
					]
				  }
				}
			  },
			"errors": []
		}`, id)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(out))
	}))

	return server
}

func DummyErrorAPI() *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		out := `{
			"data": null,
			"errors": [
				{
				  "message": "generated: load_balancer not found",
				  "path": [
					"loadBalancer"
				  ]
				}
			  ],
		}`

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(out))
	}))
}
