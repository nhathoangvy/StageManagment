package main

func Worker() {
	connect, channel, err := ConnectMQ()
	defer connect.Close()
	defer channel.Close()

	q, err := channel.QueueDeclare("", false, false, true, false, nil,) // name, durable, delete when usused, exclusive, no-wait, arguments
	failOnError(err, "Failed to declare a queue")

	err = channel.QueueBind(q.Name, KEYROUTING, EXCHANGE, false, nil) // queue name, routing key, exchange
	failOnError(err, "Failed to bind a queue")

	msgs, err := channel.Consume(q.Name, "", true, false, false, false, nil,) // queue, consumer, auto ack, exclusive, no local, no wait, args
	failOnError(err, "Failed to register a consumer")

  always := make(chan bool)
  go Receiver(msgs)
  <-always
}