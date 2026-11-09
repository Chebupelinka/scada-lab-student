FROM debian:trixie-20261104-slim AS build
RUN apt-get update \
 && apt-get install -y --no-install-recommends g++ cmake make pkg-config libmodbus-dev \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY CMakeLists.txt .
COPY src ./src
RUN cmake -S . -B build && cmake --build build

FROM debian:trixie-20261104-slim
RUN apt-get update \
 && apt-get install -y --no-install-recommends libmodbus5 \
 && rm -rf /var/lib/apt/lists/*
COPY --from=build /src/build/plc-emulator /usr/local/bin/plc-emulator
EXPOSE 5020
ENTRYPOINT ["plc-emulator"]
