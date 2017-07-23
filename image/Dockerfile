FROM ubuntu:latest
RUN apt update
RUN apt upgrade -y
ADD server .
RUN mkdir data
RUN touch data/text