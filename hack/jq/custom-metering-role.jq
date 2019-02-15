# Input ***REMOVED***le is the role we're creating multiple copies of.
. as $role
# $METERING_OPERATOR_TARGET_NAMESPACES is a comma separated list of namespaces
# to create this role in. Falls back to $METERING_NAMESPACE.
| $ENV.METERING_OPERATOR_TARGET_NAMESPACES // $ENV.METERING_NAMESPACE
# Turn the env var into a list of strings.
| split(",")
|
{
    apiVersion: "rbac.authorization.k8s.io/v1",
    kind: "RoleList",
    # For each namespace, create a copy of the role.
    items: map(
        # Each namespace from the list is passed to map.
        . as $namespace
        # The base role going into the list.
        | $role
        # Update the role's name to be pre***REMOVED***xed with our namespace,
        # in case other metering-operators are targeting the namespace.
        | .metadata.name = $namespace + "-" + .metadata.name
        # Update the role's namespace.
        | .metadata.namespace = $namespace
    )
}
