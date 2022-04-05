# arcade-token-machine
An implementation of a special-case of the Home Depot "arcade" token dispensor.

"Arcade" is a tool used by the Home Depot clouddriver-in-go implementaion to provide tokens
to connect to varuous Kubernetes clusters.  This is a slightly modified API that will
provide the Spinnaker account name as well as the provider, making it posssible to
provide many cloud providers with different credentials per Kubernetes cluster.

Home Depot seems to use one single account within a cloud to provide tokens for any
cluster, while other customers of ours have many accounts doing this, rather than
one all-powerful cloud-provider-level account.

The provider here is named "multipass" for lack of a more clever name.
