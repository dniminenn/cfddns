[![Build and Release](https://github.com/dniminenn/cfddns/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/dniminenn/cfddns/actions/workflows/build.yml)

# CFDDNS - A Simple Dynamic DNS Updater

Welcome to **CFDDNS**, a lightweight and easy-to-use dynamic DNS updater written in Go. If you've got a dynamic IP address and need to keep your DNS records up-to-date across multiple providers, you're in the right place!

## Features

- **Multi-Provider Support**: Update DNS records on Cloudflare, AWS Route53, DigitalOcean DNS, Google Cloud DNS, and DuckDNS.
- **IPv4 and IPv6 Support**: Handles both IPv4 and IPv6 addresses seamlessly.
- **Connectivity Check**: Verify internet connectivity before updating DNS records, and update records automatically when the connection is restored.
- **Customizable Intervals**: Set how often you want to check for IP changes.
- **Flexible Operation**: Run it once, keep it running as a daemon, or schedule it with cron.
- **Easy Configuration**: Simple YAML file to set up your preferences and credentials.

## Table of Contents

- [Getting Started](#getting-started)
  - [Download Binary](#download-binary)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Configuration](#configuration)
  - [General Settings](#general-settings)
  - [Provider Settings](#provider-settings)
    - [Cloudflare](#cloudflare)
    - [AWS Route53](#aws-route53)
    - [DigitalOcean](#digitalocean)
    - [Google Cloud DNS](#google-cloud-dns)
    - [DuckDNS](#duckdns)
    - [No-IP](#no-ip)
    - [FreeDNS](#freedns)
- [Usage](#usage)
  - [Running Once](#running-once)
  - [Running as a Daemon](#running-as-a-daemon)
  - [Systemd Service](#systemd-service)
  - [FreeBSD Service](#freebsd-service)
  - [SysV Init Service](#sysv-init-service)
  - [Cron Job](#cron-job)
- [Contributing](#contributing)
- [License](#license)

## Getting Started

### Download Binary

You can download the precompiled binary directly to avoid the hassle of compiling from source.

1. **Download the Binary**

   Head over to the [Releases](https://github.com/dniminenn/cfddns/releases/tag/v0.4) page and download the binary suitable for your operating system:

   - **For Linux (64-bit)**:

     ```bash
     wget https://github.com/dniminenn/cfddns/releases/download/v0.4/cfddns_v0.4_linux_amd64.tar.gz
     tar -xzf cfddns_v0.4_linux_amd64.tar.gz
     ```

   - **For macOS (64-bit)**:

     ```bash
     curl -LO https://github.com/dniminenn/cfddns/releases/download/v0.4/cfddns_v0.4_darwin_amd64.tar.gz
     tar -xzf cfddns_v0.4_darwin_amd64.tar.gz
     ```

   - **For FreeBSD (64-bit)**:

     ```bash
     fetch https://github.com/dniminenn/cfddns/releases/download/v0.4/cfddns_v0.4_freebsd_amd64.tar.gz
     tar -xzf cfddns_v0.4_freebsd_amd64.tar.gz
     ```

   - **For Windows (64-bit)**:

     Download the Windows ZIP file from [Releases](https://github.com/dniminenn/cfddns/releases/tag/v0.4).

2. **Make the Binary Executable**

   ```bash
   chmod +x cfddns
   ```

3. **(Optional) Move the Binary to Your PATH**

   ```bash
   sudo mv cfddns /usr/local/bin/
   ```

   Now you can run `cfddns` from anywhere!

### Prerequisites

- **None!** If you're using the precompiled binary, you're all set. 🎉

  _Note: If you prefer building from source, make sure you have Go 1.16 or higher installed. You can get it from the [official website](https://golang.org/dl/)._

### Installation

If you'd rather build CFDDNS yourself, follow these steps:

1. **Clone the Repository**

   ```bash
   git clone https://github.com/dniminenn/cfddns.git
   cd cfddns
   ```

2. **Build the Executable**

   ```bash
   go build -o cfddns
   ```

3. **(Optional) Move the Executable to Your PATH**

   ```bash
   sudo mv cfddns /usr/local/bin/
   ```

   And you're good to go!

## Configuration

CFDDNS uses a YAML configuration file to manage settings. By default, it looks for `cfddns.yml` in the following locations:

- Environment variable `CFDDNS_CONFIG_PATH`
- `$HOME/.config/cfddns/cfddns.yml`
- `/etc/cfddns/cfddns.yml`
- Current directory

There is a sample configuration file name `cfddns.example.yml` in the repository. You can copy it to one of the locations above and modify it to suit your needs.

### General Settings

Here's a breakdown of the general settings you can configure:

```yaml
generalSettings:
  updateInterval: 300                # Time in seconds between IP checks
  connectivityCheckInterval: 10      # Time in seconds between connectivity checks
  connectivityCheckIP: "1.1.1.1"     # IP used to check internet connectivity
  connectivityCheckPort: "53"        # Port used for connectivity check
```

- **updateInterval**: How often (in seconds) to check for IP address changes.
- **connectivityCheckInterval**: How often (in seconds) to check for internet connectivity.
- **connectivityCheckIP** and **connectivityCheckPort**: The IP and port used to verify internet access.

### Provider Settings

You can configure multiple providers under the `providers` section. Here's how to set up each supported provider:

#### Cloudflare

```yaml
providers:
  - type: "cloudflare"
    settings:
      zone: "example.com"
      apiToken: "your_cloudflare_api_token"
    records:
      - name: "home.example.com"
        type: "A"
        proxied: false
        ttl: 300
      - name: "home.example.com"
        type: "AAAA"
        proxied: false
        ttl: 300
```

- **zone**: Your domain name managed in Cloudflare.
- **apiToken**: Cloudflare API token with appropriate permissions.

#### AWS Route53

```yaml
providers:
  - type: "route53"
    settings:
      zone: "example.com"
      region: "us-east-1"
      accessKeyId: "your_aws_access_key_id"
      secretAccessKey: "your_aws_secret_access_key"
    records:
      - name: "home.example.com"
        type: "A"
        ttl: 300
```

- **accessKeyId** and **secretAccessKey**: Your AWS credentials with permissions to modify Route53 records.
- **region**: AWS region (default is `us-east-1`).
- **zone**: Your Route53 hosted zone domain.

#### DigitalOcean

```yaml
providers:
  - type: "digitalocean"
    settings:
      apiToken: "your_digitalocean_api_token"
      domain: "example.com"
    records:
      - name: "home"
        type: "A"
        ttl: 300
```

- **apiToken**: Your DigitalOcean API token.
- **domain**: Your domain name registered with DigitalOcean.

#### Google Cloud DNS

```yaml
providers:
  - type: "clouddns"
    settings:
      projectId: "your_gcp_project_id"
      credentialsJsonPath: "/path/to/credentials.json"
      zone: "example-com"
    records:
      - name: "home.example.com"
        type: "A"
        ttl: 300
```

- **projectId**: Your Google Cloud project ID.
- **credentialsJsonPath**: Path to your GCP service account credentials JSON file.
- **zone**: DNS zone name in Cloud DNS.

#### DuckDNS

Note: DuckDNS does not have an official API, so CFDDNS uses the DuckDNS update URL to update records. As a result, DuckDNS does not support the creation of new records. You must create the records manually on the DuckDNS website.

```yaml
providers:
  - type: "duckdns"
    settings:
      token: "your_duckdns_token"
    records:
      - name: "subdomain"
        type: "A"
```

- **token**: Your DuckDNS token.
- **domain**: Your DuckDNS subdomain.

#### No-IP

CFDDNS supports updating DNS records with No-IP, a popular dynamic DNS provider.

```yaml

providers:
  - type: "noip"
    settings:
      username: "your_noip_username"
      password: "your_noip_password"
    records:
      - name: "yourhostname.no-ip.org"
        type: "A"
```

- **username**: Your No-IP account username.
- **password**: Your No-IP account password.
- **name**: The hostname you have registered with No-IP.

Note: Ensure that you have registered the hostname on the No-IP website before configuring CFDDNS.

#### FreeDNS

CFDDNS also supports FreeDNS, a free dynamic DNS service.

```yaml

providers:
  - type: "freedns"
    records:
      - name: "yourhostname.mooo.com"
        type: "A"
        updateToken: "your-freedns-update-token"
```

- **name**: The full hostname you have registered with FreeDNS.
- **updateToken**: Your FreeDNS update token for the specific hostname.

Notes:

- You must create the hostname on the FreeDNS website before configuring CFDDNS.
- Each hostname in FreeDNS has a unique update token found in the "Dynamic DNS" section of your FreeDNS account.

## Usage

### Running Once

To update your DNS records immediately:

```bash
./cfddns
```

### Running as a Daemon

To keep CFDDNS running in the background and update records at intervals:

```bash
./cfddns -daemon
```

- **Verbose Mode**: Add `-verbose` to get more detailed logs.

### Systemd Service

You can set up CFDDNS as a systemd service for automatic startup and management on systems that use **systemd** (e.g., Ubuntu, Fedora).

1. **Create a Service File**

   Create a file named `cfddns.service` in `/etc/systemd/system/` with the following content:

   ```ini
   [Unit]
   Description=CFDDNS Service
   After=network.target

   [Service]
   Type=simple
   ExecStart=/usr/local/bin/cfddns -daemon
   Restart=on-failure

   [Install]
   WantedBy=multi-user.target
   ```

2. **Reload systemd and Start the Service**

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl start cfddns.service
   sudo systemctl enable cfddns.service
   ```

3. **Check Service Status**

   ```bash
   sudo systemctl status cfddns.service
   ```

### FreeBSD Service

For FreeBSD systems, you can set up CFDDNS to run as a service using the `rc.d` system.

1. **Create an rc.d Script**

   Create a script named `cfddns` in `/usr/local/etc/rc.d/` with the following content:

   ```sh
   #!/bin/sh
   #
   # PROVIDE: cfddns
   # REQUIRE: NETWORKING
   # KEYWORD: shutdown

   . /etc/rc.subr

   name="cfddns"
   rcvar=cfddns_enable
   pidfile="/var/run/${name}.pid"
   command="/usr/local/bin/${name}"
   command_args="-daemon"
   start_cmd="${name}_start"
   stop_cmd="${name}_stop"

   cfddns_start() {
       echo "Starting ${name}..."
       ${command} ${command_args} &
       echo $! > ${pidfile}
   }

   cfddns_stop() {
       echo "Stopping ${name}..."
       kill `cat ${pidfile}`
       rm -f ${pidfile}
   }

   load_rc_config $name
   run_rc_command "$1"
   ```

2. **Make the Script Executable**

   ```bash
   chmod +x /usr/local/etc/rc.d/cfddns
   ```

3. **Enable the Service**

   Add the following line to `/etc/rc.conf`:

   ```sh
   cfddns_enable="YES"
   ```

4. **Start the Service**

   ```bash
   service cfddns start
   ```

5. **Check Service Status**

   ```bash
   service cfddns status
   ```

### SysV Init Service

For systems that use **SysV init** (e.g., older versions of Debian, CentOS 6), you can create an init script.

1. **Create an Init Script**

   Create a script named `cfddns` in `/etc/init.d/` with the following content:

   ```sh
   #!/bin/sh
   #
   # chkconfig: 2345 99 10
   # description: CFDDNS Service

   ### BEGIN INIT INFO
   # Provides:          cfddns
   # Required-Start:    $network
   # Required-Stop:     $network
   # Default-Start:     2 3 4 5
   # Default-Stop:      0 1 6
   # Short-Description: CFDDNS Service
   ### END INIT INFO

   NAME="cfddns"
   DAEMON="/usr/local/bin/${NAME}"
   PIDFILE="/var/run/${NAME}.pid"
   DAEMON_OPTS="-daemon"

   start() {
       echo "Starting $NAME..."
       $DAEMON $DAEMON_OPTS &
       echo $! > $PIDFILE
   }

   stop() {
       echo "Stopping $NAME..."
       kill `cat $PIDFILE`
       rm -f $PIDFILE
   }

   case "$1" in
       start)
           start
           ;;
       stop)
           stop
           ;;
       restart)
           stop
           start
           ;;
       status)
           if [ -f $PIDFILE ]; then
               echo "$NAME is running."
           else
               echo "$NAME is stopped."
           fi
           ;;
       *)
           echo "Usage: $0 {start|stop|restart|status}"
           exit 1
   esac
   exit 0
   ```

2. **Make the Script Executable**

   ```bash
   chmod +x /etc/init.d/cfddns
   ```

3. **Add the Service to Startup**

   - On **Debian/Ubuntu**:

     ```bash
     update-rc.d cfddns defaults
     ```

   - On **CentOS**:

     ```bash
     chkconfig --add cfddns
     ```

4. **Start the Service**

   ```bash
   service cfddns start
   ```

5. **Check Service Status**

   ```bash
   service cfddns status
   ```

### Cron Job

If you prefer not to run CFDDNS as a daemon or service, you can use a cron job to schedule it to run at regular intervals. This method will execute CFDDNS at specified times, but please note that you will **lose the automatic restore on connectivity restored** feature, since CFDDNS won't be running continuously to monitor connectivity.

#### Setting Up a Cron Job

1. **Edit the Crontab**

   Open your crontab file using the following command:

   ```bash
   crontab -e
   ```

2. **Add the Cron Job Entry**

   Add the following line to schedule CFDDNS to run every 5 minutes (adjust the interval as needed):

   ```cron
   */5 * * * * /usr/local/bin/cfddns >> /var/log/cfddns.log 2>&1
   ```

   - **Explanation**:
     - `*/5 * * * *`: Runs every 5 minutes.
     - `/usr/local/bin/cfddns`: Path to the CFDDNS executable.
     - `>> /var/log/cfddns.log 2>&1`: Redirects output and errors to a log file.

3. **Save and Exit**

   Save the file and exit the editor. The cron job is now set up.

#### Important Notes

- **Loss of Connectivity Monitoring**: When using a cron job, CFDDNS won't be able to detect when internet connectivity is restored and update your DNS records immediately. It will only run at the scheduled times.
- **Log Rotation**: Over time, the log file `/var/log/cfddns.log` can grow large. Consider setting up log rotation or modify the cron job to prevent the log from growing indefinitely.

   For example, to prevent logging:

   ```cron
   */5 * * * * /usr/local/bin/cfddns >/dev/null 2>&1
   ```

#### Example Crontab Entry

Here's a sample crontab entry that runs CFDDNS every hour without logging:

```cron
0 * * * * /usr/local/bin/cfddns >/dev/null 2>&1
```

- This will run CFDDNS at the top of every hour.

## Contributing

Feel free to fork the project, make your changes, and submit a pull request! Whether it's a bug fix, new feature, or documentation improvement, contributions are welcome.

## License

This project is licensed under the MIT License.
