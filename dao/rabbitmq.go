package dao

import (
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"im_socket_server/config"
	"im_socket_server/logs"
	"sync"
	"time"
)

type MsgText struct {
	MsgContents string
}

// 实现发送者
func (t *MsgText) MsgContent() string {
	return t.MsgContents
}

// 实现接收者
func (t *MsgText) Consumer(dataByte []byte) error {
	fmt.Println(string(dataByte))
	return nil
}

// 定义全局变量,指针类型
var mqConn *amqp.Connection
var mqChan *amqp.Channel

// 定义生产者接口
type Producer interface {
	MsgContent() string
}

// 定义接收者接口
type Receiver interface {
	Consumer([]byte) error
}

// 定义RabbitMQ对象
type RabbitMQ struct {
	connection   *amqp.Connection
	channel      *amqp.Channel
	queueName    string // 队列名称
	routingKey   string // key名称
	exchangeName string // 交换机名称
	exchangeType string // 交换机类型
	producerList []Producer
	receiverList []Receiver
	mu           sync.RWMutex
}

// 定义队列交换机对象
type QueueExchange struct {
	QuName string // 队列名称
	RtKey  string // key值
	ExName string // 交换机名称
	ExType string // 交换机类型
}

// 链接rabbitMQ
func (r *RabbitMQ) mqConnect() error {
	c := config.GetConfig()
	host := c.GetString("rabbitmq.host")
	port := c.GetInt("rabbitmq.port")
	username := c.GetString("rabbitmq.username")
	password := c.GetString("rabbitmq.password")
	var err error
	RabbitUrl := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, port)
	fmt.Println(RabbitUrl)
	mqConn, err = amqp.Dial(RabbitUrl)
	r.connection = mqConn // 赋值给RabbitMQ对象
	if err != nil {
		logs.Loggers.Error("MQ打开链接失败:", zap.Error(err))
		return err
	}
	mqChan, err = mqConn.Channel()
	if err != nil {
		logs.Loggers.Error("MQ打开管道失败:", zap.Error(err))
		return err
	}
	if mqChan != nil {
		r.channel = mqChan // 赋值给RabbitMQ对象
	}
	return err
}

// 关闭RabbitMQ连接
func (r *RabbitMQ) mqClose() {
	// 先关闭管道,再关闭链接
	err := r.channel.Close()
	if err != nil {
		logs.Loggers.Error("MQ管道关闭失败:", zap.Error(err))
	}
	err = r.connection.Close()
	if err != nil {
		logs.Loggers.Error("MQ链接关闭失败:%s \n", zap.Error(err))
	}
}

// 创建一个新的操作对象
func New(q *QueueExchange) *RabbitMQ {
	return &RabbitMQ{
		queueName:    q.QuName,
		routingKey:   q.RtKey,
		exchangeName: q.ExName,
		exchangeType: q.ExType,
	}
}

// 启动RabbitMQ客户端,并初始化
func (r *RabbitMQ) Start() {
	// 开启监听生产者发送任务
	for _, producer := range r.producerList {
		go r.listenProducer(producer)
	}
	// 开启监听接收者接收任务
	for _, receiver := range r.receiverList {
		go r.listenReceiver(receiver)
	}
	time.Sleep(1 * time.Second)
}

// 注册发送指定队列指定路由的生产者
func (r *RabbitMQ) RegisterProducer(producer Producer) {
	logs.Loggers.Info("RegisterProducer", zap.String("mq", "send"))
	r.producerList = append(r.producerList, producer)
}

// 发送任务
func (r *RabbitMQ) listenProducer(producer Producer) {
	// 处理结束关闭链接
	defer r.mqClose()
	// 验证链接是否正常,否则重新链接
	if r.channel == nil {
		err := r.mqConnect()
		if err != nil {
			logs.Loggers.Error("监听接收者接收任务-ERROR:", zap.Error(err))
			return
		}
	}
	// 用于检查队列是否存在,已经存在不需要重复声明
	_, err := r.channel.QueueDeclarePassive(r.queueName, true, false, false, true, nil)
	if err != nil {
		// 队列不存在,声明队列
		// name:队列名称;durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;noWait:是否非阻塞,
		// true为是,不等待RMQ返回信息;args:参数,传nil即可;exclusive:是否设置排他
		_, err = r.channel.QueueDeclare(r.queueName, true, false, false, true, nil)
		if err != nil {
			fmt.Printf("MQ注册队列失败:%s \n", err)
			return
		}
	}
	// 队列绑定
	err = r.channel.QueueBind(r.queueName, r.routingKey, r.exchangeName, true, nil)
	if err != nil {
		fmt.Printf("MQ绑定队列失败:%s \n", err)
		return
	}
	// 用于检查交换机是否存在,已经存在不需要重复声明
	err = r.channel.ExchangeDeclarePassive(r.exchangeName, r.exchangeType, true, false, false, true, nil)
	if err != nil {
		// 注册交换机
		// name:交换机名称,kind:交换机类型,durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;
		// noWait:是否非阻塞, true为是,不等待RMQ返回信息;args:参数,传nil即可; internal:是否为内部
		err = r.channel.ExchangeDeclare(r.exchangeName, r.exchangeType, true, false, false, true, nil)
		if err != nil {
			fmt.Printf("MQ注册交换机失败:%s \n", err)
			return
		}
	}
	// 发送任务消息
	err = r.channel.Publish(r.exchangeName, r.routingKey, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(producer.MsgContent()),
	})
	if err != nil {
		fmt.Printf("MQ任务发送失败:%s \n", err)
		return
	}
}

// 注册接收指定队列指定路由的数据接收者
func (r *RabbitMQ) RegisterReceiver(receiver Receiver) {
	r.mu.Lock()
	r.receiverList = append(r.receiverList, receiver)
	fmt.Println("RegisterReceiver")
	r.mu.Unlock()
}

// 监听接收者接收任务
func (r *RabbitMQ) listenReceiver(receiver Receiver) {
	// 处理结束关闭链接
	defer r.mqClose()
	// 验证链接是否正常
	if r.channel == nil {
		err := r.mqConnect()
		if err != nil {
			logs.Loggers.Error("监听接收者接收任务-ERROR:", zap.Error(err))
			return
		}
	}
	// 用于检查队列是否存在,已经存在不需要重复声明
	_, err := r.channel.QueueDeclarePassive(r.queueName, true, false, false, true, nil)
	if err != nil {
		// 队列不存在,声明队列
		// name:队列名称;durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;noWait:是否非阻塞,
		// true为是,不等待RMQ返回信息;args:参数,传nil即可;exclusive:是否设置排他
		_, err = r.channel.QueueDeclare(r.queueName, true, false, false, true, nil)
		if err != nil {
			fmt.Printf("MQ注册队列失败:%s \n", err)
			return
		}
	}
	// 绑定任务
	err = r.channel.QueueBind(r.queueName, r.routingKey, r.exchangeName, true, nil)
	if err != nil {
		fmt.Printf("绑定队列失败:%s \n", err)
		return
	}
	// 获取消费通道,确保rabbitMQ一个一个发送消息
	err = r.channel.Qos(1, 0, true)
	msgList, err := r.channel.Consume(r.queueName, "", false, false, false, false, nil)
	if err != nil {
		fmt.Printf("获取消费通道异常:%s \n", err)
		return
	}
	for msg := range msgList {
		// 处理数据
		err := receiver.Consumer(msg.Body)
		if err != nil {
			err = msg.Ack(true)
			if err != nil {
				fmt.Printf("确认消息未完成异常:%s \n", err)
				return
			}
		} else {
			// 确认消息,必须为false
			err = msg.Ack(false)
			if err != nil {
				fmt.Printf("确认消息完成异常:%s \n", err)
				return
			}
			return
		}
	}
}
func (r *RabbitMQ) StartAMQPConsume() {
	defer func() {
		if err := recover(); err != nil {
			time.Sleep(3 * time.Second)
			fmt.Println("休息3秒")
			r.StartAMQPConsume()
		}
	}()
	if r.channel == nil {
		err := r.mqConnect()
		if err != nil {
			logs.Loggers.Error("监听接收者接收任务-ERROR:", zap.Error(err))
			return
		}
	}
	defer r.channel.Close()
	closeChan := make(chan *amqp.Error, 1)
	notifyClose := r.channel.NotifyClose(closeChan) //一旦消费者的channel有错误，产生一个amqp.Error，channel监听并捕捉到这个错误
	closeFlag := false
	msgs, err := r.channel.Consume(
		"test1",
		"",
		true,
		false,
		false,
		false, nil)
	if err != nil {
		logs.Loggers.Error("StartAMQPConsume-err", zap.Error(err))
	}
	for {
		select {
		case e := <-notifyClose:
			fmt.Println("chan通道错误,e:%s", e.Error())
			close(closeChan)
			time.Sleep(5 * time.Second)
			r.StartAMQPConsume()
			closeFlag = true
		case msg := <-msgs:
			fmt.Println("msg:", string(msg.Body))
		}
		if closeFlag {
			break
		}
	}
}
