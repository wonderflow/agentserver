# 碰到的一些问题&&要做没做的功能

* 之前cloudagent的internal设置的时间很短：3s,导致nats收到大量消息，根本发布过来，
  估计是ruby 并发量大以后，导致数据发送不出去，tcp连接丢失，所以opentsdb那页没收到数据
  目前设置到30s。这本agentserver则是60s检查一次，测试环境。到时候实际可以调整到3600s

* 如何最方便的确定一个节点是正常移除还是意外丢失连接？解决的方法是设置api。
  帮用户转发一个请求，去etcd建立一个/remove目录，设置临时节点。
  临时节点的ttl为internal两倍的node，表示要去掉的节点，然后每次检查这个目录，不检查目录中的节点是否存活。

* 【使用方法二解决了】core number 没收进metric。这就导致system.cpu.load这个数据实际上没法真的设置阈值。解决方法：
  - ssh到每个alive的host，然后`cat /proc/cpuinfo | grep processor | sort | uniq | wc -l`
  - 在cloudagent里面加这个字段，在healthmonitor里面加这个字段的处理，发送到opentsdb。需要修改另外两个系统。

* 【待解决】metric为process的时候，需要加上tag，如：{job:dea_next,index:0}，然后需要获得url上相应的tag设置。
  同时，在连接opentsdb的时候，修改相应的tags，不能在 metric_limit_map 里面直接加，这个需要处理。

* 【待解决】app相关的信息，监控都没做，衍生的功能是：应用访问大的时候调整实例数量

* 【待解决】router.requests 没数据，估计是bug

* 【配置数据持久化】存到文件，然后再读出来

* 【日志信息】输出日志 /log/agentserver.go文件下

* 【监控进程服务化】server xx start/restart/stop

—— 孙健波 2015.8.9