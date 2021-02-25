package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	cloudwatchhook "github.com/josh-hogle/logrus-cloudwatch-hook"
	"github.com/sirupsen/logrus"
)

// app constants
const (
	AppName = "go-vanity-server"
)

// global variables
var (
	log         *logrus.Entry
	Version     string
	Build       string
	ReleaseDate string
)

func main() {
	// parse command-line flags
	cloudWatchGroupName := flag.String("cloudwatch-group", "/aws/lambda/go-vanity-server",
		"Amazon CloudWatch log group name")
	cloudWatchStreamName := flag.String("cloudwatch-stream", "default", "Amazon CloudWatch log stream name")
	cloudWatchRegion := flag.String("cloudwatch-region", "", "Amazon CloudWatch region in which to store logs")
	cloudWatchRetentionDays := flag.Int("cloudwatch-retention-days", 3, "Number of days to retain logs in Amazon "+
		"CloudWatch; only used if cloudwatch-group does not exist")
	cloudWatchKmsKeyID := flag.String("cloudwatch-kms-key-id", "", "KMS key ID to use for encrypting Amazon "+
		"CloudWatch logs; only used if cloudwatch-group does not exist")
	cloudWatchTags := flag.String("cloudwatch-tags", "", "Tags to add to Amazon CloudWatch log group in the form "+
		"key1=value1,key2=value2,...; only used if cloudwatch-group does not exist")
	cloudWatchBatchFrequency := flag.String("cloudwatch-batch-frequency", "0", "Duration to wait before sending any "+
		"queued log messages to Amazon CloudWatch")
	debugFlag := flag.Bool("debug", false, "Enable debug logging")
	versionFlag := flag.Bool("version", false, "Show version information")
	flag.Parse()
	visitedFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		visitedFlags[f.Name] = true
	})

	// show version and exit
	if *versionFlag {
		fmt.Printf("%s version %s, build %s (Released %s)\n", AppName, Version, Build, ReleaseDate)
		os.Exit(0)
	}

	// check environment for settings
	if _, isSet := visitedFlags["cloudwatch-group"]; !isSet {
		if val, isSet := os.LookupEnv("CLOUDWATCH_GROUP"); isSet {
			*cloudWatchGroupName = val
		}
	}
	if _, isSet := visitedFlags["cloudwatch-stream"]; !isSet {
		if val, isSet := os.LookupEnv("CLOUDWATCH_STREAM"); isSet {
			*cloudWatchStreamName = val
		}
	}
	if _, isSet := visitedFlags["cloudwatch-region"]; !isSet {
		if val, isSet := os.LookupEnv("CLOUDWATCH_REGION"); isSet {
			*cloudWatchRegion = val
		}
	}
	if _, isSet := visitedFlags["cloudwatch-retention-days"]; !isSet {
		if val, isSet := os.LookupEnv("CLOUDWATCH_RETENTION_DAYS"); isSet {
			days, err := strconv.Atoi(val)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: 'CLOUDWATCH_RETENTION_DAYS' is not a valid number.")
				os.Exit(1)
			}
			*cloudWatchRetentionDays = days
		}
	}
	if _, isSet := visitedFlags["cloudwatch-kms-key-id"]; !isSet {
		if val, isSet := os.LookupEnv("CLOUDWATCH_KMS_KEY_ID"); isSet {
			*cloudWatchKmsKeyID = val
		}
	}
	if _, isSet := visitedFlags["cloudwatch-tags"]; !isSet {
		if val, isSet := os.LookupEnv("CLOUDWATCH_TAGS"); isSet {
			*cloudWatchTags = val
		}
	}
	if _, isSet := visitedFlags["cloudwatch-batch-frequency"]; !isSet {
		if val, isSet := os.LookupEnv("CLOUDWATCH_BATCH_FREQUENCY"); isSet {
			*cloudWatchBatchFrequency = val
		}
	}
	if _, isSet := visitedFlags["debug"]; !isSet {
		if val, isSet := os.LookupEnv("DEBUG"); isSet {
			val = strings.ToLower(val)
			if val == "1" || val == "true" || val == "enable" || val == "enabled" || val == "yes" || val == "on" {
				*debugFlag = true
			} else {
				*debugFlag = false
			}
		}
	}

	// validate input
	batchFrequency, err := time.ParseDuration(*cloudWatchBatchFrequency)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: 'cloudwatch-batch-frequncy' must be a valid time duration")
		os.Exit(1)
	}

	// configure the global package logger we'll use
	// - we aren't using the logrus global since we will add fields to this logger later\
	hook, err := cloudwatchhook.NewCloudWatchLogsHook(*cloudWatchRegion, *cloudWatchGroupName, *cloudWatchStreamName,
		cloudwatchhook.WithGroupRetentionDays(*cloudWatchRetentionDays),
		cloudwatchhook.WithBatchDuration(batchFrequency),
		cloudwatchhook.WithGroupKmsKeyID(*cloudWatchKmsKeyID),
		/*WithTags()*/
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to create log hook: %v", err)
		os.Exit(2)
	}
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.Hooks.Add(hook)
	logger.Out = ioutil.Discard
	if *debugFlag {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	// log a message
	logger.Info("This is a test message")
	logger.WithFields(logrus.Fields{
		"field1": "value1",
	}).Error("This is an error")

	os.Exit(0)
}
