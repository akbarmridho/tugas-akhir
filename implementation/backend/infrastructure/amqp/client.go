package amqp

import (
	"context"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"tugas-akhir/backend/infrastructure/amqp/entity"
	"tugas-akhir/backend/pkg/utility"

	"sync"
	"time"
)

// Client is the base struct for handling connection recovery, consumption and
// publishing. Note that this struct has an internal mutex to safeguard against
// data races.
type Client struct {
	M               *sync.Mutex
	Queue           *entity.QueueConfig
	Exchange        *entity.ExchangeConfig
	Logger          *zap.Logger
	connection      *amqp091.Connection
	channel         *amqp091.Channel
	done            chan bool
	notifyConnClose chan *amqp091.Error
	notifyChanClose chan *amqp091.Error
	notifyConfirm   chan amqp091.Confirmation
	isReady         bool
}

func (client *Client) WaitUntilReady(ctx context.Context) {
	for {
		client.M.Lock()

		if client.isReady {
			client.M.Unlock()
			break
		}

		client.M.Unlock()

		utility.SleepWithContext(ctx, time.Second)
	}
}

// handleReconnect will wait for a connection error on
// notifyConnClose, and then continuously attempt to reconnect.
func (client *Client) handleReconnect(address string) {
	for {
		client.M.Lock()
		client.isReady = false
		client.M.Unlock()

		client.Logger.Info("attempting to connect")

		conn, err := client.connect(address)
		if err != nil {
			client.Logger.Error("failed to connect. Retrying...")

			select {
			case <-client.done:
				return
			case <-time.After(entity.ReconnectDelay):
			}
			continue
		}

		if done := client.handleReInit(conn); done {
			break
		}
	}
}

// connect will create a new AMQP connection
func (client *Client) connect(address string) (*amqp091.Connection, error) {
	conn, err := amqp091.Dial(address)
	if err != nil {
		return nil, err
	}

	client.changeConnection(conn)
	client.Logger.Info("connected")
	return conn, nil
}

// handleReInit will wait for a channel error
// and then continuously attempt to re-initialize both channels
func (client *Client) handleReInit(conn *amqp091.Connection) bool {
	for {
		client.M.Lock()
		client.isReady = false
		client.M.Unlock()

		err := client.init(conn)
		if err != nil {
			client.Logger.Info("failed to initialize channel, retrying...")

			select {
			case <-client.done:
				return true
			case <-client.notifyConnClose:
				client.Logger.Info("connection closed, reconnecting...")
				return false
			case <-time.After(entity.ReInitDelay):
			}
			continue
		}

		select {
		case <-client.done:
			return true
		case <-client.notifyConnClose:
			client.Logger.Info("connection closed, reconnecting...")
			return false
		case <-client.notifyChanClose:
			client.Logger.Info("channel closed, re-running init...")
		}
	}
}

// init will initialize channel & declare queue
func (client *Client) init(conn *amqp091.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	err = ch.Confirm(false)
	if err != nil {
		return err
	}

	if client.Exchange != nil {
		err = ch.ExchangeDeclare(
			client.Exchange.Name,
			client.Exchange.Kind,
			client.Exchange.Durable,
			client.Exchange.AutoDelete,
			client.Exchange.Internal,
			client.Exchange.NoWait,
			nil,
		)

		if err != nil {
			return err
		}
	}

	if client.Queue != nil {
		args := amqp091.Table{}

		if client.Queue.Timeout != nil {
			args[amqp091.ConsumerTimeoutArg] = client.Queue.Timeout.Milliseconds()
		}

		_, err = ch.QueueDeclare(
			client.Queue.Name,
			client.Queue.Durable,    // Durable
			client.Queue.AutoDelete, // Delete when unused
			client.Queue.Exclusive,  // Exclusive
			client.Queue.NoWait,     // No-wait
			args,                    // Arguments
		)
	}

	if err != nil {
		return err
	}

	client.changeChannel(ch)
	client.M.Lock()
	client.isReady = true
	client.M.Unlock()
	client.Logger.Info("client init done")

	return nil
}

// changeConnection takes a new connection to the queue,
// and updates the close listener to reflect this.
func (client *Client) changeConnection(connection *amqp091.Connection) {
	client.connection = connection
	client.notifyConnClose = make(chan *amqp091.Error, 1)
	client.connection.NotifyClose(client.notifyConnClose)
}

// changeChannel takes a new channel to the queue,
// and updates the channel listeners to reflect this.
func (client *Client) changeChannel(channel *amqp091.Channel) {
	client.channel = channel
	client.notifyChanClose = make(chan *amqp091.Error, 1)
	client.notifyConfirm = make(chan amqp091.Confirmation, 1)
	client.channel.NotifyClose(client.notifyChanClose)
	client.channel.NotifyPublish(client.notifyConfirm)
}

// Close will cleanly shut down the channel and connection.
func (client *Client) Close() error {
	client.M.Lock()
	// we read and write isReady in two locations, so we grab the lock and hold onto
	// it until we are finished
	defer client.M.Unlock()

	if !client.isReady {
		return entity.AlreadyClosedError
	}
	close(client.done)
	err := client.channel.Close()
	if err != nil {
		return err
	}
	err = client.connection.Close()
	if err != nil {
		return err
	}

	client.isReady = false
	return nil
}
