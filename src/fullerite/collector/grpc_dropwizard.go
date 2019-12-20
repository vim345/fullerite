package collector

// import "git@git.yelpcorp.com:services/ranking_platform_query_engine.git/generated-go/yelpcorp.com/ranking_platform/query_engine/metrics"
import "yelpcorp.com/query_engine_go_clientlib/metrics"

func say() string {
	m := metrics.MetricRequest{}
	return m.XXX_size_cache
	// return "Hello"
}
