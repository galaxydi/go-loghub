package producer

import (
	"fmt"
	"os"
	"testing"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
)

func TestVsSign(t *testing.T) {
	producerConfig := GetDefaultProducerConfig()
	producerConfig.Endpoint = os.Getenv("LOG_TEST_ENDPOINT")
	provider := sls.NewStaticCredentialsProvider(os.Getenv("LOG_TEST_ACCESS_KEY_ID"), os.Getenv("LOG_TEST_ACCESS_KEY_SECRET"), "")
	producerConfig.CredentialsProvider = provider
	producerConfig.Region = os.Getenv("LOG_TEST_REGION")
	producerConfig.AuthVersion = sls.AuthV4
	producerInstance := InitProducer(producerConfig)

	producerInstance.Start() // 启动producer实例
	for i := 0; i < 100; i++ {
		// GenerateLog  is producer's function for generating SLS format logs
		log := GenerateLog(uint32(time.Now().Unix()), map[string]string{"content": "test", "content2": fmt.Sprintf("%v", i)})
		err := producerInstance.SendLog(os.Getenv("LOG_TEST_PROJECT"), os.Getenv("LOG_TEST_LOGSTORE"), "127.0.0.1", "topic", log)
		if err != nil {
			fmt.Println(err)
		}
	}
	producerInstance.Close(60)   // 有限关闭，传递int值，参数值需为正整数，单位为秒
	producerInstance.SafeClose() // 安全关闭
}
