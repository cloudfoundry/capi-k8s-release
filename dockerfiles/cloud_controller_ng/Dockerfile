FROM ruby:2.5.5
WORKDIR /cloud_controller_ng

RUN apt-get update && apt-get install -y git \
  libxml2-dev \
  libxslt-dev \
  libmariadb-dev \
  tzdata \
  libpq-dev \
  jq \
  less \
  dnsutils

COPY cloud_controller_ng/Gemfile .
COPY cloud_controller_ng/Gemfile.lock .

RUN bundle config build.nokogiri --use-system-libraries
RUN bundle install

COPY cloud_controller_ng/ .
COPY lifecycle_hooks/ .

ENV BUNDLE_GEMFILE /cloud_controller_ng/Gemfile
ENV CLOUD_CONTROLLER_NG_CONFIG /config/cloud_controller_ng.yml
ENV C_INCLUDE_PATH /libpq/include
ENV DYNO #{spec.job.name}-#{spec.index}
ENV LANG en_US.UTF-8
ENV LIBRARY_PATH /libpq/lib
ENV RAILS_ENV production

ENTRYPOINT ["/cloud_controller_ng/bin/cloud_controller", "-c", "/config/cloud_controller_ng.yml"]
