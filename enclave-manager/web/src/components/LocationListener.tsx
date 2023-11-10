import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useKurtosisClient } from "../client/enclaveManager/KurtosisClientContext";

export const LocationListener = () => {
  const client = useKurtosisClient();
  const navigate = useNavigate();

  useEffect(() => {
    if (client.getCloudUrl()) {
      console.log(client.getParentRequestedRoute());
      const route = client.getParentRequestedRoute();
      if (route) navigate(route);
    }
  }, [client.getCloudUrl()]);

  return <></>;
};
