FROM ubuntu:16.04

# Pick up some TF dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential \
        curl \
        libfreetype6-dev \
        libpng12-dev \
        libzmq3-dev \
        pkg-config \
        python \
        rsync \
        software-properties-common \
        unzip \
        && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install pip
RUN curl -O https://bootstrap.pypa.io/get-pip.py && \
    python3 get-pip.py && \
    rm get-pip.py

# Install tensorflow and its related package
RUN pip3 --no-cache-dir install \
        Pillow \
        numpy \
        scipy \
        tensorflow

WORKDIR /app

COPY ./bin /app

ENTRYPOINT ["sh", "./nerual-style-data-server"]

# --- DO NOT EDIT OR DELETE BETWEEN THE LINES --- #
# These lines will be edited automatically by parameterized_docker_build.sh. #
# COPY _PIP_FILE_ /
# RUN pip --no-cache-dir install /_PIP_FILE_
# RUN rm -f /_PIP_FILE_

# Install TensorFlow CPU version from central repo
#RUN pip3 --no-cache-dir install tensorflow
# --- ~ DO NOT EDIT OR DELETE BETWEEN THE LINES --- #

# TensorBoard
#EXPOSE 6006
# IPython
#EXPOSE 8888