#! /bin/bash

installSoftware() {
    apt -qq -y install nginx
}

installMyAnime() {
    mkdir -p /var/www/myanime
    curl -Lo- https://github.com/sunshineplan/myanime/releases/download/v1.0/release.tar.gz | tar zxC /var/www/myanime
    cd /var/www/myanime
    chmod +x myanime
}

configMyAnime() {
    read -p 'Please enter API address: ' api
    read -p 'Please enter unix socket(default: /run/myanime.sock): ' unix
    [ -z $unix ] && unix=/run/myanime.sock
    read -p 'Please enter host(default: 127.0.0.1): ' host
    [ -z $host ] && host=127.0.0.1
    read -p 'Please enter port(default: 12345): ' port
    [ -z $port ] && port=12345
    read -p 'Please enter log path(default: /var/log/app/myanime.log): ' log
    [ -z $log ] && log=/var/log/app/myanime.log
    read -p 'Please enter update URL: ' update
    read -p 'Please enter exclude files: ' exclude
    mkdir -p $(dirname $log)
    sed "s,\$api,$api," /var/www/myanime/config.ini.default > /var/www/myanime/config.ini
    sed -i "s,\$unix,$unix," /var/www/myanime/config.ini
    sed -i "s,\$log,$log," /var/www/myanime/config.ini
    sed -i "s/\$host/$host/" /var/www/myanime/config.ini
    sed -i "s/\$port/$port/" /var/www/myanime/config.ini
    sed -i "s,\$update,$update," /var/www/myanime/config.ini
    sed -i "s|\$exclude|$exclude|" /var/www/myanime/config.ini
    ./myanime install || exit 1
    service myanime start
}

writeLogrotateScrip() {
    if [ ! -f '/etc/logrotate.d/app' ]; then
	cat >/etc/logrotate.d/app <<-EOF
		/var/log/app/*.log {
		    copytruncate
		    rotate 12
		    compress
		    delaycompress
		    missingok
		    notifempty
		}
		EOF
    fi
}

setupNGINX() {
    cp -s /var/www/myanime/scripts/myanime.conf /etc/nginx/conf.d
    sed -i "s/\$domain/$domain/" /var/www/myanime/scripts/myanime.conf
    service nginx reload
}

main() {
    read -p 'Please enter domain:' domain
    installSoftware
    installMyAnime
    configMyAnime
    writeLogrotateScrip
    setupNGINX
}

main
