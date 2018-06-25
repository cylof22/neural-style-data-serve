package NSUtil

import (
	"os"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
)

// Register add the service to the consul register center
func Register(consulAddress, consulPort, advertiseAddress, advertisePort string,
	serviceName, tag string, log log.Logger) (registar sd.Registrar) {
	// Service discovery domain. In this example we use Consul.
	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		consulConfig.Address = consulAddress + ":" + consulPort
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			log.Log("err", err)
			os.Exit(1)
		}
		client = consulsd.NewClient(consulClient)
	}

	check := api.AgentServiceCheck{
		HTTP:     "http://" + advertiseAddress + ":" + advertisePort + "/health",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Basic health checks",
	}

	port, _ := strconv.Atoi(advertisePort)
	serviceID := serviceName + "-" + advertiseAddress + ":" + advertisePort
	asr := api.AgentServiceRegistration{
		ID:      serviceID, //unique service ID
		Name:    serviceName,
		Address: advertiseAddress,
		Port:    port,
		Tags:    []string{tag},
		Check:   &check,
	}
	return consulsd.NewRegistrar(client, &asr, log)
}
