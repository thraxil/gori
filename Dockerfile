FROM ubuntu
MAINTAINER Anders Pearson <anders@columbia.edu>
ADD ./gori /usr/local/bin/gori
ADD ./media /var/www/gori/media
ENTRYPOINT ["/usr/local/bin/gori"]
EXPOSE 8888
