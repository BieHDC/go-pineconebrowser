# go-pineconebrowser
This project implements web serving over [Pinecone](https://github.com/matrix-org/pinecone) which is used for [Matrix Dendrite P2P](https://github.com/matrix-org/dendrite/tree/main/cmd/dendrite-demo-pinecone).
You can imagine it like a **very bare bones** I2P or Tor where we tunnel http over a non standard transport.
It is more meant as a proof of concept than something you should deploy, but you can!

---
## Crash Course
There are two variants. Both of them by default spawn a Gui at [http://127.0.0.1:4200](http://127.0.0.1:4200) which lists all the connected Pinecone instances and whether they are a "Pinehost" Instance or a normal one.

If they are a Pinehost it displays their name, if they are a Webhoster, and which endpoints they support (for future expansion). If they are a Webhoster, their UUID (public key hex as string) will be clickable and you browse to their instance.

As for hosting, the hosting end is just a proxy connecting to a Webserver. This can be some locally hosted website in the style of reverse proxy or any random website. Keep in mind there is zero effort made to prevent something like deanonymisation!

It goes like: Your Webbrowser -> Pinehost -> Pinehost -> Webserver. And then the response in reverse.

Also in this first release it only does local lan multicast discovery and does not connect to any fixed peers somewhere. So your friend somewhere else wont find your instance, but you can [trivially add](https://github.com/matrix-org/dendrite/blob/main/cmd/dendrite-demo-pinecone/main.go#L90) that option if you want. It is kept this way so someone messing around does not unintentionally mess around with others.

---
## Variants
- [cmd/pineproxy](/cmd/pineproxy/): This implements a forwarding http_proxy as you know from Tor when you setup your browser proxy settings. You then also have the UUID in the URL Bar instead of the temporary proxy address, making the links shareable for example.
    - Pro: The right way to do it. Use this if in doubt.
    - Con: Needs browser settings to be applied. (Set http_proxy and make sure to disable https-only mode)

- [cmd/pineweb](/cmd/pineweb/): This works by spawning a temporary local webserver when you browse to a Pinehost which then proxies all the requests to the remote Pinehost and the responses back.
    - Pro: No browser proxy setting required.
    - Con: Hacky at best and is likely to be unstable.

---
## Ideas
- Use the public key to validate the remote host and do a TOFU (Trust On First Use) database and send a challange to prevent impersonation.
- Have self declared nice urls like in I2P: verylongaddress.b32.i2p -> yourname.i2p which then can also be secured through TOFU above.
- Suggest something!

---
## TODOs
- Also do a https_proxy for accessability reasons.
- As mentioned in [pinehost_endpoints.go:25](/pinehost_endpoints.go#L25) for the "/webhoster" endpoint, it could return something useful like what this Instance hosts, maybe tags about the content etc. And add this to the gui. Right now its only used to check if there is a webhost going on.

---
## FIXMEs
- Anything not wrapped inside a HTTP interface has a high fail rate and i dont yet know why. This is why the Pinehost interface is http instead of something more low level. Good enough for this proof of concept.
- In the [first contact](/pinehost.go#L97) and [also here](/pinehost.go#L141) we have the problem that the connection seems very unstable. The two issues are that the instances dont like to be contacted on the same thing at the very same time, hence the random sleep, and also not immediatly after they have been added.

---
## License
AGPLv3
