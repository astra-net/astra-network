# Build
FROM golang:1.16 AS build

RUN apt update -y && \
    apt install libgmp-dev libssl-dev git -y && \
    apt -y clean all

ENV GOPATH=/root/go
ENV GO111MODULE=on
ENV ASTRA_PATH=${GOPATH}/src/github.com/astra-net
RUN mkdir -p $ASTRA_PATH

WORKDIR $ASTRA_PATH

RUN git clone https://github.com/astra-net/astra-network.git && \
    git clone https://github.com/astra-net/bls.git && \
    git clone https://github.com/astra-net/mcl.git

WORKDIR $ASTRA_PATH/astra

RUN go mod download

RUN go mod tidy

RUN make linux_static && \
    cp ./bin/astra /root/astra && \
    cp ./rosetta/infra/run.sh /root/run.sh && \
    cp ./rosetta/infra/rclone.conf /root/rclone.conf

RUN cp ./rosetta/infra/astra-pstn.conf /root/astra-pstn.conf && \
    cp ./rosetta/infra/astra-mainnet.conf /root/astra-mainnet.conf && \
    cp ./.astra/rosetta_local_fix.csv /root/rosetta_local_fix.csv

# Execution
FROM ubuntu:20.04

RUN apt update -y && \
    apt install libgmp-dev libssl-dev ca-certificates rclone -y && \
    apt -y clean all \

WORKDIR /root

COPY --from=build /root/astra /root/astra
COPY --from=build /root/run.sh /root/run.sh
COPY --from=build /root/rclone.conf /root/.config/rclone/rclone.conf
COPY --from=build /root/astra-pstn.conf /root/astra-pstn.conf
COPY --from=build /root/astra-mainnet.conf /root/astra-mainnet.conf
COPY --from=build /root/rosetta_local_fix.csv /root/rosetta_local_fix.csv

EXPOSE 9500/tcp
EXPOSE 9700/tcp
EXPOSE 9800/tcp
ENTRYPOINT ["/bin/bash","/root/run.sh"]
