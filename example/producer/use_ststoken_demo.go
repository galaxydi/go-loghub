package main

import (
	"fmt"
	"github.com/aliyun/aliyun-log-go-sdk/producer"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	producerConfig := producer.GetDefaultProducerConfig()
	producerConfig.Endpoint = os.Getenv("Endpoint")
	producerConfig.AccessKeyID = os.Getenv("AccessKeyID")
	producerConfig.AccessKeySecret = os.Getenv("AccessKeySecret")
	//When the producer is closed, if the StsTokenShutDown parameter is not set to nil, it will actively call the close method to close the channel.
	producerConfig.StsTokenShutDown = make(chan struct{})
	producerConfig.UpdateStsToken = updateStsToken
	producerInstance := producer.InitProducer(producerConfig)
	ch := make(chan os.Signal)
	signal.Notify(ch)
	producerInstance.Start()
	var m sync.WaitGroup
	for i := 0; i < 10; i++ {
		m.Add(1)
		go func() {
			defer m.Done()
			for i := 0; i < 1000; i++ {
				// GenerateLog  is producer's function for generating SLS format logs
				// GenerateLog has low performance, and native Log interface is the best choice for high performance.
				log := producer.GenerateLog(uint32(time.Now().Unix()), map[string]string{"content": "test", "content2": fmt.Sprintf("%v", i)})
				err := producerInstance.SendLog("project", "logstrore", "topic", "127.0.0.1", log)
				if err != nil {
					fmt.Println(err)
				}
			}
		}()
	}
	m.Wait()
	fmt.Println("Send completion")
	if _, ok := <-ch; ok {
		fmt.Println("Get the shutdown signal and start to shut down")
		producerInstance.Close(60)
	}
}

func updateStsToken() (accessKeyID, accessKeySecret, securityToken string, expireTime time.Time, err error) {
	// 写入自己的获取的 ststoken和过期时间逻辑代码，producer会自动在ststoken到达过期时间的时候，重新执行该函数去获取最新的ststoken以及其过期时间。
	// TODO 此处填入自己的获取ststoken 的逻辑
	return accessKeyID, accessKeySecret, securityToken, expireTime, nil

}