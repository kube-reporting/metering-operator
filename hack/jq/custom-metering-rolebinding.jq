# Input ***REMOVED***le is the role binding we're creating multiple copies of.
. as $rolebinding
# $METERING_OPERATOR_TARGET_NAMESPACES is a comma separated list of namespaces
# to create this rolebinding in. Falls back to $METERING_NAMESPACE.
| $ENV.METERING_OPERATOR_TARGET_NAMESPACES // $ENV.METERING_NAMESPACE
# Turn the env var into a list of strings.
| split(",")
|
{
    apiVersion: "rbac.authorization.k8s.io/v1",
    kind: "RoleBindingList",
    # For each namespace, create a copy of the rolebinding
    items: map(
        # Each namespace from the list is passed to map.
        . as $namespace
        # The base rolebinding going into the list.
        | $rolebinding
        # Update the rolebinding's name to be pre***REMOVED***xed with our namespace,
        # in case other metering-operators are targeting the namespace.
        | .metadata.name = $namespace + "-" + .metadata.name
        # Update the rolebinding's namespace.
        | .metadata.namespace = $namespace
        # Update the roleRef to match our rolebinding name. The role and
        # rolebinding should have the same name.
        | .roleRef.name = .metadata.name
        # Update the service account to the one where the
        # metering-operator is running.
        | .subjects[0].namespace=$ENV.METERING_NAMESPACE
    )
}
