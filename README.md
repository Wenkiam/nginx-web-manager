一个简单的nginx配置web管理程序，提供了配置管理和证书管理两个功能。[示例网站](https://nwm-demo.starcloud.cc) 密码 123456

### 配置管理
配置管理提供了配置修改和nginx reload功能，可以在页面上修改指定目录下（默认是/etc/nginx/conf.d，可在页面修改）的nginx配置并reload nginx使配置生效。

### 证书管理
证书管理功能基于[lego](https://go-acme.github.io/lego/)的ACME库实现，可以在页面上选择dns提供商，配置好凭证信息后即可申请指定域名的Let's Encrypt证书，可以指定证书存放目录，同时提供了下载功能。
可以在页面配置证书到期检查定时任务，定时任务会在证书到期前3天自动重新申请，也可以手动在页面点击续期。


### 使用方法
下载对应平台的可执行程序，在命令行运行。linux用户提供了service方式运行，执行install.sh脚本即可。提供了和nginx一起在docker中运行的dockerfile，该镜像基于
nginx docker 官方镜像。也可以直接从docker hub下载：https://hub.docker.com/repository/docker/wenkiam/nginx/general

#### 查看帮助信息
```shell
nwm --help
NAME:
   nwm - A web server to manage nginx and certificates

USAGE:
   nwm [global options] command [command options] [arguments...]

COMMANDS:
   start    Start a web server to manage nginx and certificates
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)

```

```shell
nwm start --help

NAME:
   nwm start - Start a web server to manage nginx and certificates

USAGE:
   nwm start [command options] [arguments...]

OPTIONS:
   --container value       if your nginx is running in docker mode,set the nginx container name to this value [$CONTAINER]
   --cron value            cron expression for renew check (default: 0 0 0 * * ?) [$CRON]
   --dns value             Name of dns provider [$DNS_PROVIDER]
   --email value           Email used for cert registration and recovery contact [$EMAIL]
   --log value             log directory [$NWM_LOG]
   --nginx.conf value      Directory of nginx config files (default: "/etc/nginx/conf.d/") [$NGINX_CONF]
   --oidc.client value     clientId of oauth [$OIDC_CLIENT]
   --oidc.secret value     Client Secret of oauth [$OIDC_SECRET]
   --oidc.site value       site of oauth server [$OIDC_SITE]
   --password value        password to login system [$AUTH_PASSWORD]
   --path value            Directory to use for storing the data. (default: ".") [$CERT_PATH]
   --port value, -p value  port of web server (default: 8080) [$PORT]
   --url value             CA hostname (and optionally :port). The server certificate must be trusted in order to avoid further modifications to the client.(default: https://acme-v02.api.letsencrypt.org/directory) [$CA_URL]

```
### 手动启动示例
nwm start --path /etc/nginx/ssl --log /logs

### 参数说明
程序运行参数可以在命令行指定，也可以通过环境变量指定，对应的环境变量可以通过帮助命令查看

#### --nginx.conf
这个参数用于指定nginx 配置文件所在目录，默认是/etc/nginx/conf.d，可以在页面上随时调整

#### --container
如果你的nginx运行在docker上，而管理程序运行在宿主机上，需要通过这个参数指定nginx的容器名，同时nginx配置目录应该配置为宿主机挂载目录

#### --port
web服务端口，默认8080

#### --email
用于证书申请的身份凭证之一

#### url
ACME服务器目录URL，默认值 https://acme-v02.api.letsencrypt.org/directory ，通常不需要动

#### dns
你的域名解析服务商，基本主流的域名解析服务商都支持，不同的域名提供商需要的凭证信息不同，这些凭证信息可以配置在环境变量里，也可以在服务启动后在页面管理里面配置。这些域名解析商需要的环境变量详情可查看：[https://go-acme.github.io/lego/dns/](https://go-acme.github.io/lego/dns/)

#### --path
证书存放目录，申请证书时会在该目录下建立以域名作为文件夹名的文件夹，申请下来的证书会放到对应的文件夹里

#### --log
日志存放目录，如果不配置则日志默认输出到控制台

#### --cron
证书过期检查定时任务cron表达式，默认 0 0 0 * * ?，即每天0点执行

### 登录认证
v0.0.2版本新增了登录认证功能，支持简单密码登录和通过oidc登录，如果启动时没有配置相关参数，则不启用认证功能，如果同时设置了简单密码和oidc相关参数，则只会开启oidc认证。相关参数说明如下所示

#### --password
系统登录密码

#### --oidc.client
ClientID 的值

#### --oidc.secret
ClientSecret 的值

#### --oidc.site
认证服务器地址，以keycloak为例，这个值为 https://your_keycloak_hostname/realms/your_realm
