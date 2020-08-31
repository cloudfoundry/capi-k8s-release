FROM cloudfoundry/build:base

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        wget \
        gnupg \
        software-properties-common && \
    # install postgresql-client
    sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt bionic-pgdg main" > /etc/apt/sources.list.d/pgdg.list' && \
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && \
    apt-get update && \
    apt-get install -y --no-install-recommends postgresql-client && \
    # install yq
    add-apt-repository -y ppa:rmescandon/yq && \
    apt-get update && \
    apt-get install -y --no-install-recommends yq && \
    # clean up
    apt-get purge -y \
        wget \
        gnupg \
        software-properties-common && \
    apt-get clean && \
    apt-get autoclean && \
    rm -rf /var/lib/apt/lists/*

# backup metadata
ADD backup-metadata /bin/backup-metadata

