#!/bin/bash
file=nwm
if [ ! -f $file ]; then
    go build -o $file
fi
dir=/usr/local/nginx-web-manager
mkdir -p $dir
cp $file $dir
ln -s $dir/$file /usr/bin/
serviceFile=/usr/lib/systemd/system/nwm.service
if [ ! -f $serviceFile ]; then
    touch $serviceFile
fi

cat > $serviceFile <<- EOF
[Unit]
Description=Nginx Web Manager
After=network.target
SuccessAction=none

[Service]
Type=simple
Restart=on-failure
RestartSec=5s
ExecStart=$dir/$file start --path /etc/nginx/ssl --log $dir

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable nwm
systemctl restart nwm

echo 'install nginx web manager success'