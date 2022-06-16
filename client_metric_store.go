package sls

import "time"

// CreateMetricStore .
func (c *Client) CreateMetricStore(project, name string, ttl, shard int) error {
	logStore := &LogStore{
		Name:          name,
		TTL:           ttl,
		ShardCount:    shard,
		TelemetryType: "Metrics",
		AutoSplit:     true,
		MaxSplitShard: 64,
	}
	err := c.CreateLogStoreV2(project, logStore)
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 3)
	subStore := &SubStore{}
	subStore.Name = "prom"
	subStore.SortedKeyCount = 2
	subStore.TimeIndex = 2
	subStore.TTL = ttl
	subStore.Keys = append(subStore.Keys, SubStoreKey{
		Name: "__name__",
		Type: "text",
	}, SubStoreKey{
		Name: "__labels__",
		Type: "text",
	}, SubStoreKey{
		Name: "__time_nano__",
		Type: "long",
	}, SubStoreKey{
		Name: "__value__",
		Type: "double",
	})
	if !subStore.IsValid() {
		panic("metric store invalid")
	}
	return c.CreateSubStore(project, name, subStore)
}

// UpdateMetricStore .
func (c *Client) UpdateMetricStore(project, name string, ttl int) error {
	metricStore := &LogStore{
		Name:          name,
		TelemetryType: "Metrics",
		TTL:           ttl,
	}
	err := c.UpdateLogStoreV2(project, metricStore)
	if err != nil {
		return err
	}
	return c.UpdateSubStoreTTL(project, name, ttl)
}

// DeleteMetricStore .
func (c *Client) DeleteMetricStore(project, name string) error {
	return c.DeleteLogStore(project, name)
}

// GetMetricStore .
func (c *Client) GetMetricStore(project, name string) (*LogStore, error) {
	return c.GetLogStore(project, name)
}
