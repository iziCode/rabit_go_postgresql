package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"math/rand"
	"net/http"
)

func main() {
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/send", sendMsg)

	fmt.Println("starting server at :8081")
	http.ListenAndServe(":8081", nil)

}

func mainPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	w.Write(uploadFormTmpl)
}

var uploadFormTmpl = []byte(`
<html>
	<body>
	<form action="/send" method="post" enctype="multipart/form-data">
		Your Messages: <input type="text" name="my_text">
		<input type="submit" value="Send">
	</form>
	</body>
</html>
`)

func sendMsg(w http.ResponseWriter, r *http.Request) {
	formValue := r.FormValue("my_text")

	fmt.Println("formValue", formValue)
	rabbitAndWorkersIsOk, statusCodeCheck := checkRabbitAndWorkers()

	if rabbitAndWorkersIsOk {

		res, isOk, statusCode := runRPC(formValue)
		fmt.Println(res, isOk, statusCode)
		if isOk {
			w.WriteHeader(statusCode)
			fmt.Fprintf(w, "Выполнено\n")
			fmt.Fprintf(w, "Status %v", statusCode)
		} else {
			w.WriteHeader(statusCode)
			fmt.Fprintf(w, "Ошибка! Не выполнено!\n")
			fmt.Fprintf(w, "Status %v", statusCode)
		}

	} else {
		if statusCodeCheck == 503 {
			w.WriteHeader(statusCodeCheck)
			fmt.Fprintf(w, "Воркеры не запущены!\n")
			fmt.Fprintf(w, "Status %v", statusCodeCheck)
		} else {
			w.WriteHeader(statusCodeCheck)
			fmt.Fprintf(w, "Ошибка! Не выполнено!\n")
			fmt.Fprintf(w, "Status %v", statusCodeCheck)
		}
	}

}
func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func runRPC(n string) (res string, isOK bool, statusCode int) {
	url := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Println("failed to open a amqp.Dial. Somethings wrong with RabbitMQ. URL: "+url+" Error:", err)
		return res, false, http.StatusInternalServerError
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Println("failed to open a channel. Somethings wrong with RabbitMQ. Error:", err)
		return res, false, http.StatusInternalServerError
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Println("failed to declare a queue. Somethings wrong with RabbitMQ. Error:", err)
		return res, false, http.StatusInternalServerError
	}
	fmt.Printf("queue %s have %d msg and %d consumers\n",
		q.Name, q.Messages, q.Consumers)

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println("failed to register a consumer. Somethings wrong with RabbitMQ. Error:", err)
		return res, false, http.StatusInternalServerError
	}

	corrId := randomString(32)

	err = ch.Publish(
		"",
		"test",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          []byte(n),
		})
	if err != nil {
		log.Println("failed to publish a message. Somethings wrong with RabbitMQ. Error:", err)
		return res, false, http.StatusInternalServerError
	}

	for d := range msgs {
		if corrId == d.CorrelationId {
			res = string(d.Body)
			if res != "200" {
				return res, false, http.StatusServiceUnavailable
			}
			break
		}
	}

	return res, true, http.StatusOK
}

func checkRabbitAndWorkers() (bool, int) {
	url := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Println("failed to open a amqp.Dial. Somethings wrong with RabbitMQ. URL: "+url+" Error:", err)
		return false, http.StatusInternalServerError
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Println("failed to open a channel. Somethings wrong with RabbitMQ. Error:", err)
		return false, http.StatusInternalServerError
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"test",
		false,
		false,
		false,
		false,
		amqp.Table{"passive": "true"},
	)

	if err != nil {
		log.Println("failed to declare a queue. Somethings wrong with RabbitMQ. Error:", err)

		return false, http.StatusInternalServerError
	}

	if q.Consumers > 0 {
		return true, http.StatusOK
	} else {
		log.Println("No workers")
		return false, http.StatusServiceUnavailable
	}
}
