package collect

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
	troubleshootv1beta1 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta1"
)

type RedisOutput map[string][]byte

func Redis(ctx *Context, databaseCollector *troubleshootv1beta1.Database) ([]byte, error) {
	databaseConnection := DatabaseConnection{}

	opt, err := redis.ParseURL(databaseCollector.URI)
	if err != nil {
		databaseConnection.Error = err.Error()
	} else {
		client := redis.NewClient(opt)
		stringResult := client.Info("server")

		if stringResult.Err() != nil {
			databaseConnection.Error = stringResult.Err().Error()
		}

		databaseConnection.IsConnected = stringResult.Err() == nil

		if databaseConnection.Error == "" {
			lines := strings.Split(stringResult.Val(), "\n")
			for _, line := range lines {
				lineParts := strings.Split(line, ":")
				if len(lineParts) == 2 {
					if lineParts[0] == "redis_version" {
						databaseConnection.Version = strings.TrimSpace(lineParts[1])
					}
				}
			}
		}
	}

	b, err := json.Marshal(databaseConnection)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal database connection")
	}

	collectorName := databaseCollector.CollectorName
	if collectorName == "" {
		collectorName = "redis"
	}

	redisOutput := map[string][]byte{
		fmt.Sprintf("redis/%s.json", collectorName): b,
	}

	bb, err := json.Marshal(redisOutput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal redis output")
	}

	return bb, nil
}
