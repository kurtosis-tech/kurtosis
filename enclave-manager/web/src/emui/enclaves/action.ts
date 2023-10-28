import { ActionFunction, json, redirect } from "react-router-dom";
import { KurtosisClient } from "../../client/enclaveManager/KurtosisClient";
import { isDefined } from "../../utils";

export const enclavesAction =
  (kurtosisClient: KurtosisClient): ActionFunction =>
  async ({ params, request }) => {
    const formData = await request.json();
    const intent = formData["intent"];
    if (intent === "delete") {
      const uuids = formData["enclaveUUIDs"];
      if (!isDefined(uuids)) {
        throw json({ message: "Missing enclaveUUIDs" }, { status: 400 });
      }
      console.log(uuids);
      await Promise.all(uuids.map((uuid: string) => kurtosisClient.destroy(uuid)));
      return redirect("/enclaves");
    } else {
      console.log("blep");
      throw json({ message: "Invalid intent" }, { status: 400 });
    }
  };
