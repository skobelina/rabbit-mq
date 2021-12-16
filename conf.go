package rabbit

type Configuration struct {
	AMQPConnectionURL string
}

var Config = Configuration{
	AMQPConnectionURL: "amqp://guest:guest@localhost:5672/",
}
