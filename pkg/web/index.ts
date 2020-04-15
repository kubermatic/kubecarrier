// import { APIVersion, VersionRequest } from "../apis/apiserver/v1alpha1/kubecarrier_pb";
// import {KubecarrierClient} from "../apis/apiserver/v1alpha1/KubecarrierServiceClientPb";

import {VersionRequest } from "./kubecarrier_pb";
import {KubecarrierClient } from "./KubecarrierServiceClientPb";

console.log("I'm alive and working...more  or less")
let request = new VersionRequest()
let client = new KubecarrierClient("http://" + window.location.hostname + ":8090", null, null)
client.version(request, {}, ((err, response) => {
    console.log("get message back v3")
    if (err) {
        console.error(err.code, err.message)
    }
    console.log(`version: ${response.getVersion()}`)
}));
