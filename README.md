This project is a component of the [Operator Framework](https://github.com/operator-framework), an open source toolkit to manage Kubernetes native applications, called Operators, in an effective, automated, and scalable way.
Read more in the [introduction blog post](https://coreos.com/blog/introducing-operator-framework-metering).

<img src="Documentation/operator_logo_metering_color.svg" height="125px"></img>

Metering records historical cluster usage, and can generate usage reports showing usage breakdowns by pod or namespace over arbitrary time periods.

## Project status: alpha

Read more about the implemented and planned features in the documentation:

 - [Installation Guide](Documentation/install-metering.md) - install Metering on your Kubernetes cluster
 - [Usage Guide](Documentation/using-metering.md) - start here to learn how to use the project
 - [Metering Architecture](Documentation/metering-architecture.md) - understand the system's components
 - [Configuration](Documentation/metering-config.md) - see the available options for talking to Prometheus, talking to AWS, storing Metering output and more.
 - [Writing Custom Reports](Documentation/writing-custom-queries.md) - extend or customize reports based on your needs

## Developers

To follow the developer getting started guide, use [Documentation/dev/developer-guide.md](Documentation/dev/developer-guide.md).
