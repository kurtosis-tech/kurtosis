import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useKurtosisClient } from "../client/enclaveManager/KurtosisClientContext";

export const LocationListener = () => {
  const client = useKurtosisClient();
  const navigate = useNavigate();
  const cloudUrl = client.getCloudUrl();

  useEffect(() => {
    if (cloudUrl) {
      const route = client.getParentRequestedRoute();
      if (route) navigate(route);
    }
  }, [cloudUrl, client, navigate]);

  return <></>;
};
