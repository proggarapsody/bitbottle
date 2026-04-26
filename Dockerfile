FROM scratch
COPY bitbottle /bitbottle
ENTRYPOINT ["/bitbottle"]
CMD ["mcp"]
