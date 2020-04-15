import {VersionRequest } from "./kubecarrier_pb";
import {KubecarrierClient } from "./KubecarrierServiceClientPb";

console.log("I'm alive and working...more  or less")
let request = new VersionRequest()
let client = new KubecarrierClient("http://" + window.location.hostname + ":8090", null, null)
client.version(request, {}, ((err, response) => {
    if (err) {
        console.error(err.code, err.message)
    } else {
        console.log(`version: ${response.getVersion()}`)
    }
}));
