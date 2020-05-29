package feeder

import (
	"encoding/csv"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/creekorful/trandoshan/internal/log"
	"github.com/creekorful/trandoshan/internal/natsutil"
	"github.com/creekorful/trandoshan/pkg/proto"
)

// GetApp return the feeder app
func GetApp() *cli.App {
	return &cli.App{
		Name:    "trandoshan-feeder",
		Version: "0.0.1",
		Usage:   "", // TODO
		Flags: []cli.Flag{
			log.GetLogFlag(),
			&cli.StringFlag{
				Name:     "nats-uri",
				Usage:    "URI to the NATS server",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "url",
				Usage:    "URL to send to the crawler",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "list",
				Usage:    "List of URLs to send to the crawler (one per line in csv file)",
				Required: false,
			},
		},
		Action: execute,
	}
}

func execute(ctx *cli.Context) error {
	log.ConfigureLogger(ctx)

	logrus.Infof("Starting trandoshan-feeder v%s", ctx.App.Version)

	logrus.Debugf("Using NATS server at: %s", ctx.String("nats-uri"))

	// Connect to the NATS server
	nc, err := nats.Connect(ctx.String("nats-uri"))
	if err != nil {
		logrus.Errorf("Error while connecting to NATS server %s: %s", ctx.String("nats-uri"), err)
		return err
	}
	defer nc.Close()

	// list of urls to crawl
	if _, err := os.Stat(ctx.String("list")); !os.IsNotExist(err) {
		file, err := os.Open(ctx.String("list"))
		if err != nil {
			return err
		}
		reader := csv.NewReader(file)
		reader.Comma = ','
		reader.LazyQuotes = true
		data, err := reader.ReadAll()
		if err != nil {
			return err
		}
		for _, link := range data {
			// Publish the message
			if err := natsutil.PublishJSON(nc, proto.URLTodoSubject, &proto.URLTodoMsg{URL: link}); err != nil {
				logrus.Errorf("Unable to publish URL: %s", err)
				return err
			}
			logrus.Infof("URL %s successfully sent to the crawler", link)
		}

	} else {
		// Publish the message
		if err := natsutil.PublishJSON(nc, proto.URLTodoSubject, &proto.URLTodoMsg{URL: ctx.String("url")}); err != nil {
			logrus.Errorf("Unable to publish URL: %s", err)
			return err
		}

		logrus.Infof("URL %s successfully sent to the crawler", ctx.String("url"))
	}

	return nil
}
