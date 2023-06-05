package sls

const (
	EVENT_STORE_TELEMETRY_TYPE = "Event"
	EVENT_STORE_INDEX          = "{\"max_text_len\":16384,\"ttl\":7,\"log_reduce\":false,\"line\":{\"caseSensitive\":false,\"chn\":true,\"token\":[\",\",\" \",\"'\",\"\\\"\",\";\",\"=\",\"(\",\")\",\"[\",\"]\",\"{\",\"}\",\"?\",\"@\",\"&\",\"<\",\">\",\"/\",\":\",\"\\n\",\"\\t\",\"\\r\"]},\"keys\":{\"specversion\":{\"type\":\"text\",\"doc_value\":true,\"alias\":\"\",\"caseSensitive\":false,\"chn\":false,\"token\":[\",\",\" \",\"'\",\"\\\"\",\";\",\"=\",\"(\",\")\",\"[\",\"]\",\"{\",\"}\",\"?\",\"@\",\"&\",\"<\",\">\",\"/\",\":\",\"\\n\",\"\\t\",\"\\r\"]},\"id\":{\"type\":\"text\",\"doc_value\":true,\"alias\":\"\",\"caseSensitive\":false,\"chn\":false,\"token\":[\",\",\" \",\"'\",\"\\\"\",\";\",\"=\",\"(\",\")\",\"[\",\"]\",\"{\",\"}\",\"?\",\"@\",\"&\",\"<\",\">\",\"/\",\":\",\"\\n\",\"\\t\",\"\\r\"]},\"source\":{\"type\":\"text\",\"doc_value\":true,\"alias\":\"\",\"caseSensitive\":false,\"chn\":false,\"token\":[\",\",\" \",\"'\",\"\\\"\",\";\",\"=\",\"(\",\")\",\"[\",\"]\",\"{\",\"}\",\"?\",\"@\",\"&\",\"<\",\">\",\"/\",\"\\n\",\"\\t\",\"\\r\"]},\"type\":{\"type\":\"text\",\"doc_value\":true,\"alias\":\"\",\"caseSensitive\":false,\"chn\":false,\"token\":[\",\",\" \",\"'\",\"\\\"\",\";\",\"=\",\"(\",\")\",\"[\",\"]\",\"{\",\"}\",\"?\",\"@\",\"&\",\"<\",\">\",\"/\",\":\",\"\\n\",\"\\t\",\"\\r\"]},\"subject\":{\"type\":\"text\",\"doc_value\":true,\"alias\":\"\",\"caseSensitive\":false,\"chn\":false,\"token\":[\",\",\" \",\"'\",\"\\\"\",\";\",\"=\",\"(\",\")\",\"[\",\"]\",\"{\",\"}\",\"?\",\"@\",\"&\",\"<\",\">\",\"/\",\":\",\"\\n\",\"\\t\",\"\\r\"]},\"datacontenttype\":{\"type\":\"text\",\"doc_value\":true,\"alias\":\"\",\"caseSensitive\":false,\"chn\":false,\"token\":[\",\",\" \",\"'\",\"\\\"\",\";\",\"=\",\"(\",\")\",\"[\",\"]\",\"{\",\"}\",\"?\",\"@\",\"&\",\"<\",\">\",\"/\",\":\",\"\\n\",\"\\t\",\"\\r\"]},\"dataschema\":{\"type\":\"text\",\"doc_value\":true,\"alias\":\"\",\"caseSensitive\":false,\"chn\":false,\"token\":[\",\",\" \",\"'\",\"\\\"\",\";\",\"=\",\"(\",\")\",\"[\",\"]\",\"{\",\"}\",\"?\",\"@\",\"&\",\"<\",\">\",\"/\",\":\",\"\\n\",\"\\t\",\"\\r\"]},\"data\":{\"type\":\"json\",\"doc_value\":true,\"alias\":\"\",\"caseSensitive\":false,\"chn\":false,\"token\":[\",\",\" \",\"'\",\"\\\"\",\";\",\"=\",\"(\",\")\",\"[\",\"]\",\"{\",\"}\",\"?\",\"@\",\"&\",\"<\",\">\",\"/\",\":\",\"\\n\",\"\\t\",\"\\r\"],\"index_all\":true,\"max_depth\":-1,\"json_keys\":{}},\"time\":{\"type\":\"text\",\"doc_value\":true,\"alias\":\"\",\"caseSensitive\":false,\"chn\":false,\"token\":[\",\",\" \",\"'\",\"\\\"\",\";\",\"=\",\"(\",\")\",\"[\",\"]\",\"{\",\"}\",\"?\",\"@\",\"&\",\"<\",\">\",\"/\",\":\",\"\\n\",\"\\t\",\"\\r\"]},\"title\":{\"type\":\"text\",\"doc_value\":true,\"alias\":\"\",\"caseSensitive\":false,\"chn\":false,\"token\":[\",\",\" \",\"'\",\"\\\"\",\";\",\"=\",\"(\",\")\",\"[\",\"]\",\"{\",\"}\",\"?\",\"@\",\"&\",\"<\",\">\",\"/\",\":\",\"\\n\",\"\\t\",\"\\r\"]},\"message\":{\"type\":\"text\",\"doc_value\":true,\"alias\":\"\",\"caseSensitive\":false,\"chn\":false,\"token\":[\",\",\" \",\"'\",\"\\\"\",\";\",\"=\",\"(\",\")\",\"[\",\"]\",\"{\",\"}\",\"?\",\"@\",\"&\",\"<\",\">\",\"/\",\":\",\"\\n\",\"\\t\",\"\\r\"]},\"status\":{\"type\":\"text\",\"doc_value\":true,\"alias\":\"\",\"caseSensitive\":false,\"chn\":false,\"token\":[\",\",\" \",\"'\",\"\\\"\",\";\",\"=\",\"(\",\")\",\"[\",\"]\",\"{\",\"}\",\"?\",\"@\",\"&\",\"<\",\">\",\"/\",\":\",\"\\n\",\"\\t\",\"\\r\"]}}}"
)

func (c *Client) CreateEventStore(project string, eventStore *LogStore) error {
	eventStore.TelemetryType = EVENT_STORE_TELEMETRY_TYPE
	err := c.CreateLogStoreV2(project, eventStore)
	if err != nil {
		return err
	}
	return c.CreateIndexString(project, eventStore.Name, EVENT_STORE_INDEX)
}

func (c *Client) UpdateEventStore(project string, eventStore *LogStore) error {
	eventStore.TelemetryType = EVENT_STORE_TELEMETRY_TYPE
	return c.UpdateLogStoreV2(project, eventStore)
}

func (c *Client) DeleteEventStore(project, name string) error {
	return c.DeleteLogStore(project, name)
}

func (c *Client) GetEventStore(project, name string) (*LogStore, error) {
	return c.GetLogStore(project, name)
}

func (c *Client) ListEventStore(project string, offset, size int) ([]string, error) {
	return c.ListLogStoreV2(project, offset, size, EVENT_STORE_TELEMETRY_TYPE)
}
