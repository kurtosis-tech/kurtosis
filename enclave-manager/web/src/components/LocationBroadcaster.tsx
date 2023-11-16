import { useEffect } from "react";
import { useLocation } from "react-router-dom";

export const LocationBroadcaster = () => {
  const location = useLocation();

  useEffect(() => {
    const message = { message: "em-ui-location-pathname", value: location.pathname };
    // eslint-disable-next-line no-restricted-globals
    parent.postMessage(message, "*");
  }, [location.pathname]);

  return <></>;
};
