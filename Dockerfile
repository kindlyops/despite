FROM        scratch
MAINTAINER  Kindly Ops, LLC <support@kindlyops.com>
ADD         bin/despite despite
ENV         DESPITE_VERBOSITY 无
ENV         DESPITE_EXIT 0
ENTRYPOINT  ["/despite"]
