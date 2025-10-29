package env

import (
	"os"
	"strconv"
)

type Envs struct {
	PROMETHEUS_PORT			   string
	BUCKET_NAME				   string
	TRANSCODE_TEMPLATE         string
	PRODUCTION_MODE            string
	APP_PORT                   string
	TEMP_CREDS_PATHS           string
	MAIN_CREDS_PATHS           string
	MONGO_URI                  string
	MONGO_USER                 string
	MONGO_PASS                 string
	CHANNEL                    string
	FFMPEG_BIN                 string
	UDP                        string
	NVIDIA_VISIBLE_DEVICES     string
	HLS_DELETE_THRESHOLD       string
	HLS_LIST_SIZE              string
	ARCHIVE_BLACKLIST_PLAYLIST string
	KAFKA_BOOTSTRAP_SERVERS    string
	KAFKA_PAYGIR_STORAGE_TOPIC string
	KAFKA_PAYGIR_MONGO_TOPIC   string
	ELASTIC_URL                string
	ELASTIC_USERNAME           string
	ELASTIC_PASSWORD           string
	ELASTIC_INDEXNAME          string
	GRPC_PORT				   string
	LOG_LEVEL				   string
	HAS_ARCHIVE                bool
}

func ReadEnvs() Envs {
	envs := Envs{}
	var err error
	envs.TRANSCODE_TEMPLATE = os.Getenv("TRANSCODE_TEMPLATE")
	envs.BUCKET_NAME = os.Getenv("BUCKET_NAME")
	envs.PRODUCTION_MODE = os.Getenv("PRODUCTION_MODE")
	envs.APP_PORT = os.Getenv("APP_PORT")
	envs.TEMP_CREDS_PATHS = os.Getenv("TEMP_CREDS_PATHS")
	envs.MAIN_CREDS_PATHS = os.Getenv("MAIN_CREDS_PATHS")
	envs.MONGO_URI = os.Getenv("MONGO_URI")
	envs.MONGO_USER = os.Getenv("MONGO_USER")
	envs.MONGO_PASS = os.Getenv("MONGO_PASS")
	envs.CHANNEL = os.Getenv("CHANNEL")
	envs.FFMPEG_BIN = os.Getenv("FFMPEG_BIN")
	envs.UDP = os.Getenv("UDP")
	envs.HLS_DELETE_THRESHOLD = os.Getenv("HLS_DELETE_THRESHOLD")
	envs.HLS_LIST_SIZE = os.Getenv("HLS_LIST_SIZE")
	envs.ARCHIVE_BLACKLIST_PLAYLIST = os.Getenv("ARCHIVE_BLACKLIST_PLAYLIST")
	envs.KAFKA_BOOTSTRAP_SERVERS = os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	envs.KAFKA_PAYGIR_STORAGE_TOPIC = os.Getenv("KAFKA_PAYGIR_STORAGE_TOPIC")
	envs.KAFKA_PAYGIR_MONGO_TOPIC = os.Getenv("KAFKA_PAYGIR_MONGO_TOPIC")
	envs.ELASTIC_URL = os.Getenv("ELASTIC_URL")
	envs.ELASTIC_USERNAME = os.Getenv("ELASTIC_USERNAME")
	envs.ELASTIC_PASSWORD = os.Getenv("ELASTIC_PASSWORD")
	envs.ELASTIC_INDEXNAME = os.Getenv("ELASTIC_INDEXNAME")
	envs.GRPC_PORT = os.Getenv("GRPC_PORT")
	envs.PROMETHEUS_PORT = os.Getenv("PROMETHEUS_PORT")
	envs.LOG_LEVEL = os.Getenv("logLevel")

	envs.HAS_ARCHIVE, err = strconv.ParseBool(os.Getenv("HAS_ARCHIVE"))
	if err != nil {
		panic("parse err HAS_ARCHIVE " + err.Error())
	}

	return envs
}
