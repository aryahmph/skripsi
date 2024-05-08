package presentations

type (
	// LogHeaderOpts Presentation
	LogHeaderOpts struct {
		Payload       interface{} `json:"payload"`
		URL           string      `json:"url"`
		Header        HeaderOpts  `json:"header"`
		Method        string      `json:"method"`
		TimeoutSecond int         `json:"timeout_second"`
	}

	// HeaderOpts presentation
	HeaderOpts struct {
		Authorization string `json:"Authorization"`
		ContentType   string `json:"Content-Type"`
	}
)
