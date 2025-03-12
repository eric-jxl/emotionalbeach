#!/bin/bash
HTML_FILE="/tmp/system_check_$(date +%Y%m%d%H%M).html"
SERVICES=("sshd" "crond" "nginx" "mysql" "docker")  # 可自定义要检查的服务

# 创建HTML文件头部
cat > "$HTML_FILE" <<EOF
<html>
<head>
<title>服务器巡检报告</title>
<meta charset="utf-8">
<style>
    body { font-family: Arial, sans-serif; margin: 20px; }
    h2 { color: #333; border-bottom: 2px solid #333; padding-bottom: 5px; }
    table { width: 100%; border-collapse: collapse; margin-bottom: 25px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
    th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
    th { background-color: #4CAF50; color: white; }
    tr:nth-child(even) { background-color: #f9f9f9; }
    pre { margin: 5px 0; padding: 10px; background-color: #f5f5f5; border: 1px solid #ddd; }
    .warning { color: #ff5722; font-weight: bold; }
</style>
</head>
<body>
<h1>服务器巡检报告</h1>
<p>生成时间：$(date "+%Y-%m-%d %H:%M:%S")</p>
EOF

# 函数：添加章节到HTML
add_section() {
    # shellcheck disable=SC2129
    echo "<h2>$1</h2>" >> "$HTML_FILE"
    echo "<table>" >> "$HTML_FILE"
    echo "$2" >> "$HTML_FILE"
    # shellcheck disable=SC2086
    echo "</table>" >> $HTML_FILE
}

# 系统信息
system_info=$(cat <<EOF
<tr><th width="25%">主机名</th><td>$(hostname)</td></tr>
<tr><th>系统版本</th><td>$(source /etc/os-release; echo "$PRETTY_NAME")</td></tr>
<tr><th>内核版本</th><td>$(uname -r)</td></tr>
<tr><th>架构</th><td>$(arch)</td></tr>
<tr><th>启动时间</th><td>$(uptime -s)</td></tr>
<tr><th>运行时间</th><td>$(uptime -p)</td></tr>
EOF
)
add_section "🖥️ 系统信息" "$system_info"

# CPU信息
cpu_info=$(lscpu | awk -F: '/Model name|Socket|Core|CPU\(s\)/ {gsub(/ +/, " ", $2); print "<tr><th>"$1"</th><td>"$2"</td></tr>"}')
add_section "⚡ CPU信息" "$cpu_info"

# 内存信息
memory_info=$(free -h | awk '/Mem|Swap/ {print "<tr><th>"$1"</th><td>"$2"</td><td>"$3"</td><td>"$4"</td></tr>"}' | sed '1i<tr><th>类型</th><th>总量</th><th>已用</th><th>剩余</th></tr>')
add_section "💾 内存信息" "$memory_info"

# 磁盘信息
disk_info=$(df -h | awk '
BEGIN {print "<tr><th>文件系统</th><th>挂载点</th><th>总大小</th><th>已用</th><th>可用</th><th>使用率</th></tr>"}
NR>1 {gsub(/\%/,""); if($5 > 80) $5="<span class=\"warning\">"$5"%</span>"; else $5=$5"%"; print "<tr><td>"$1"</td><td>"$6"</td><td>"$2"</td><td>"$3"</td><td>"$4"</td><td>"$5"</td></tr>"}')
add_section "💽 磁盘使用" "$disk_info"

# 负载信息
load_info=$(cat <<EOF
<tr><th width="25%">1分钟负载</th><td>$(awk '{print $1}' /proc/loadavg)</td></tr>
<tr><th>5分钟负载</th><td>$(awk '{print $2}' /proc/loadavg)</td></tr>
<tr><th>15分钟负载</th><td>$(awk '{print $3}' /proc/loadavg)</td></tr>
<tr><th>CPU核心数</th><td>$(nproc)</td></tr>
EOF
)
add_section "📊 系统负载" "$load_info"

# 网络信息
network_info=$(cat <<EOF
<tr><th width="25%">IP地址</th><td><pre>$(ip -br addr show | grep -v lo)</pre></td></tr>
<tr><th>路由表</th><td><pre>$(ip route | sed 's/</\&lt;/g; s/>/\&gt;/g')</pre></td></tr>
<tr><th>DNS配置</th><td><pre>$(grep nameserver /etc/resolv.conf)</pre></td></tr>
EOF
)
add_section "🌐 网络信息" "$network_info"

# 防火墙状态
fw_status=""
# 检查firewalld
if systemctl is-active firewalld &>/dev/null; then
    fw_status+="<tr><th>Firewalld 状态</th><td>$(systemctl is-active firewalld)</td></tr>"
    fw_status+="<tr><th>防火墙规则</th><td><pre>$(firewall-cmd --list-all | sed 's/</\&lt;/g; s/>/\&gt;/g')</pre></td></tr>"
# 检查ufw
elif ufw status &>/dev/null; then
    fw_status+="<tr><th>UFW 状态</th><td>$(ufw status | grep Status)</td></tr>"
    fw_status+="<tr><th>防火墙规则</th><td><pre>$(ufw status numbered | sed 's/</\&lt;/g; s/>/\&gt;/g')</pre></td></tr>"
# 检查iptables
else
    fw_status+="<tr><th>IPTables 状态</th><td>$(systemctl is-active iptables 2>/dev/null || echo '未运行')</td></tr>"
    fw_status+="<tr><th>防火墙规则</th><td><pre>$(iptables -L -n -v --line-numbers | sed 's/</\&lt;/g; s/>/\&gt;/g')</pre></td></tr>"
fi
add_section "🔥 防火墙状态" "$fw_status"

# 服务状态
service_info="<tr><th width=\"25%\">服务名称</th><th>运行状态</th><th>开机启动</th></tr>"
for service in "${SERVICES[@]}"; do
    active_status=$(systemctl is-active "$service" 2>/dev/null || echo "unknown")
    enabled_status=$(systemctl is-enabled "$service" 2>/dev/null || echo "unknown")
    [ "$active_status" != "active" ] && active_status="<span class=\"warning\">$active_status</span>"
    service_info+="<tr><td>$service</td><td>$active_status</td><td>$enabled_status</td></tr>"
done
add_section "🛎️ 服务状态" "$service_info"

# 端口监听
port_info=$(cat <<EOF
<tr><th width="25%">监听端口</th><td><pre>$(ss -tuln)</pre></td></tr>
<tr><th>TCP连接状态</th><td><pre>$(ss -s | grep TCP)</pre></td></tr>
EOF
)
add_section "🔌 网络连接" "$port_info"

# 内核参数
kernel_params=$(cat <<EOF
<tr><th width="25%">最大文件句柄数</th><td>$(sysctl -n fs.file-max)</td></tr>
<tr><th>TIME-WAIT超时</th><td>$(sysctl -n net.ipv4.tcp_fin_timeout)s</td></tr>
<tr><th>内存交换倾向</th><td>$(sysctl -n vm.swappiness)</td></tr>
<tr><th>SYN重试次数</th><td>$(sysctl -n net.ipv4.tcp_syn_retries)</td></tr>
<tr><th>最大连接数</th><td>$(sysctl -n net.core.somaxconn)</td></tr>
<tr><th>TCP快速回收</th><td>$(sysctl -n net.ipv4.tcp_tw_recycle)</td></tr>
EOF
)
add_section "⚙️ 内核参数" "$kernel_params"

# 结束HTML文件
cat >> "$HTML_FILE" <<EOF
</body>
</html>
EOF

echo "巡检报告已生成：$HTML_FILE"
