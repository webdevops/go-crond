# go-cron

[![Circle CI](https://circleci.com/gh/anarcher/go-cron.svg?style=svg)](https://circleci.com/gh/anarcher/go-cron)

# Usage

## Docker 

- https://registry.hub.docker.com/u/anarcher/go-cron/

```
$docker run anarcher/go-cron -h                            
Usage of go-cron:
    -cpu=4: maximum number of CPUs
    -file="crontab": crontab file path
```

```
$ docker run go-cron -file=crontab                                                              
2015/06/18 07:53:38 Add cron job spec:*/1 * * * * cmd:echo "hello world!" err:<nil>
2015/06/18 07:53:38 Add cron job spec:*/1 * * * * cmd:echo "hello" ; sleep 1 ; echo "world" err:<nil>
2015/06/18 07:53:38 Start runner
2015/06/18 07:53:39 cmd:echo "hello world!" out:hello world! err:<nil>
2015/06/18 07:53:40 cmd:echo "hello world!" out:hello world! err:<nil>
2015/06/18 07:53:40 cmd:echo "hello" ; sleep 1 ; echo "world" out:hello world err:<nil>
^C2015/06/18 07:53:40 Got signal:  interrupt
2015/06/18 07:53:40 Stop runner
2015/06/18 07:53:40 End cron
```

