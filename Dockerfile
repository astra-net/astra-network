FROM ubuntu:18.04

ARG TARGETARCH
ARG GOLANG_VERSION="1.16.3"

SHELL ["/bin/bash", "-c"]

ENV GOPATH=/root/go
ENV GO111MODULE=on
ENV ASTRA_PATH=${GOPATH}/src/github.com/harmony-one
ENV OPENSSL_DIR=/usr/lib/ssl
ENV MCL_DIR=${ASTRA_PATH}/mcl
ENV BLS_DIR=${ASTRA_PATH}/bls
ENV CGO_CFLAGS="-I${BLS_DIR}/include -I${MCL_DIR}/include"
ENV CGO_LDFLAGS="-L${BLS_DIR}/lib"
ENV LD_LIBRARY_PATH=${BLS_DIR}/lib:${MCL_DIR}/lib
ENV GIMME_GO_VERSION=${GOLANG_VERSION}
ENV PATH="/root/bin:${PATH}"

RUN apt update && apt upgrade -y && \
	apt install libgmp-dev libssl-dev curl git \
	psmisc dnsutils jq make gcc g++ bash tig tree sudo vim \
	silversearcher-ag unzip emacs-nox nano bash-completion -y

RUN mkdir ~/bin && \
	curl -sL -o ~/bin/gimme \
	https://raw.githubusercontent.com/travis-ci/gimme/master/gimme && \
	chmod +x ~/bin/gimme

RUN eval "$(~/bin/gimme ${GIMME_GO_VERSION})"

RUN git clone https://github.com/harmony-one/astra.git ${ASTRA_PATH}/astra

RUN git clone https://github.com/harmony-one/bls.git ${ASTRA_PATH}/bls

RUN git clone https://github.com/harmony-one/mcl.git ${ASTRA_PATH}/mcl

RUN git clone https://github.com/harmony-one/go-sdk.git ${ASTRA_PATH}/go-sdk

RUN cd ${ASTRA_PATH}/bls && make -j8 BLS_SWAP_G=1

RUN touch /root/.bash_profile && \
	gimme ${GIMME_GO_VERSION} >> /root/.bash_profile && \
	echo "GIMME_GO_VERSION='${GIMME_GO_VERSION}'" >> /root/.bash_profile && \
	echo "GO111MODULE='on'" >> /root/.bash_profile && \
	echo ". ~/.bash_profile" >> /root/.profile && \
	echo ". ~/.bash_profile" >> /root/.bashrc

ENV PATH="/root/.gimme/versions/go${GIMME_GO_VERSION}.linux.${TARGETARCH:-amd64}/bin:${GOPATH}/bin:${PATH}"

RUN . ~/.bash_profile; \
	go get -u golang.org/x/tools/cmd/goimports; \
	go get -u golang.org/x/lint/golint ; \
	go get -u github.com/rogpeppe/godef ; \
	go get -u github.com/go-delve/delve/cmd/dlv; \
	go get -u github.com/golang/mock/mockgen; \
	go get -u github.com/stamblerre/gocode; \
	go get -u golang.org/x/tools/...; \
	go get -u honnef.co/go/tools/cmd/staticcheck/...

WORKDIR ${ASTRA_PATH}/astra

RUN scripts/install_build_tools.sh

RUN go mod tidy

RUN scripts/go_executable_build.sh -S

RUN cd ${ASTRA_PATH}/go-sdk && make -j8 && cp astra /root/bin

ARG K1=0x3b00aab412891853f2cc91a6957a66088f22437b
ARG K2=0x415077f2efb732538683baf6479a94ed35b836b5
ARG K3=0xef1c0d949efbd8baed211fad28a08c5cf96e989d

ARG KS1=8d222cffa99eb1fb86c581d9dfe7d60dd40ec62aa29056b7ff48028385270541
ARG KS2=da1800da5dedf02717696675c7a7e58383aff90b1014dfa1ab5b7bd1ce3ef535
ARG KS3=f4267bb5a2f0e65b8f5792bb6992597fac2b35ebfac9885ce0f4152c451ca31a

RUN astra keys import-private-key ${KS1} && \
	astra keys import-private-key ${KS2} && \
	astra keys import-private-key ${KS3} && \
	astra keys generate-bls-key > keys.json 

RUN jq  '.["encrypted-private-key-path"]' -r keys.json > /root/keypath && cp keys.json /root && \
	echo "export BLS_KEY_PATH=$(cat /root/keypath)" >> /root/.bashrc && \
	echo "export BLS_KEY=$(jq '.["public-key"]' -r keys.json)" >> /root/.bashrc && \
	echo "printf '${K1}, ${K2}, ${K3} are imported accounts in astra for local dev\n\n'" >> /root/.bashrc && \
	echo "printf 'test with: astra blockchain validator information ${K1}\n\n'" >> /root/.bashrc && \
	echo "echo "$(jq '.["public-key"]' -r keys.json)" is an extern bls key" >> /root/.bashrc && \
	echo ". /etc/bash_completion" >> /root/.bashrc && \
	echo ". <(astra completion)" >> /root/.bashrc
