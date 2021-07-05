FROM ubuntu:16.04

WORKDIR /root
COPY turingchain turingchain
COPY turingchain-cli turingchain-cli
COPY turingchain.toml turingchain-solo.toml ./

RUN ./turingchain-cli cert --host=127.0.0.1

CMD ["/root/turingchain", "-f", "/root/turingchain.toml"]
