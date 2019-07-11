FROM alpine:latest

COPY find-failing-notifications.linux64 /usr/local/bin/find-failing-notifications
USER daemon
CMD /usr/local/bin/find-failing-notifications 
