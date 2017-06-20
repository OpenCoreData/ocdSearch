# Start from scratch image and add in a precompiled binary
# docker build  --tag="opencoredata/ocdweb:0.1"  .
# docker run -d -p 9900:9900  opencoredata/ocdweb:0.1
FROM scratch

# Add in the static elements (could also mount these from local filesystem)
ADD ocdSearch /
ADD ./static  /static
#ADD ./indexes  /indexes


# We will add in the required index directory via a local mount 
# The index is large and doesn't need to be part of the image.  
# When updated a new continer must be initiallized to open the new index(es)

# Add our binary
CMD ["/ocdSearch"]

# Document that the service listens on this port
EXPOSE 9802 9800

