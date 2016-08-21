FROM        scratch
MAINTAINER  Kindly Ops, LLC <support@kindlyops.com>
WORKDIR     /code
ADD         despite_linux_amd64 despite
ENV         DESPITE_VERBOSITY 无
ENV         DESPITE_EXIT 0
ENTRYPOINT  ["/despite"]
CMD         ["-h"]
