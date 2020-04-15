import {VersionRequest } from "./kubecarrier_pb";
import {KubecarrierClient } from "./KubecarrierServiceClientPb";
import {Empty} from "google-protobuf/google/protobuf/empty_pb";

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

console.log("let the streaming being!")
let steamReg = new Empty()
let stream = client.versionSteam(steamReg, {})

stream.on('data', function(response) {
    console.log("got steam response; version", response.getVersion(), (new Date()).getUTCMilliseconds());
});
stream.on('status', function(status) {
    console.log("steam status")
    console.log(status.code);
    console.log(status.details);
    console.log(status.metadata);
});
