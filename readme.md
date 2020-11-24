## lotus-monitor部署

* 二进制文件在99上, /home/xg/lotus-monitor
* 部署在每个miner机上
```sh
nohup ./lotus-monitor run --proxy http://ip:40001/api/v1/miner/push --interval 5m > $LOG_PATH/monitor.log &!
```