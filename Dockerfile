# 使用 golang:1.17-alpine 作为基础镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /root/

# 将页面打包进去
COPY index.html .

# 复制本地编译好的二进制文件到最终镜像中
COPY event-listener .


# 将端口暴露给外部（如果你的服务在 8080 端口上运行）
EXPOSE 8080

# 后台执行 Go 程序
CMD ["./event-listener"]