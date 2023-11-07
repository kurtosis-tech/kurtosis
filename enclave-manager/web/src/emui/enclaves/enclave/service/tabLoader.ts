import { LoaderFunctionArgs } from "react-router-dom";

export const serviceTabLoader = async ({ params }: LoaderFunctionArgs): Promise<{ routeName: string }> => {
  const activeTab = params.activeTab;

  switch (activeTab?.toLowerCase()) {
    case "overview":
      return { routeName: "Overview" };
    case "logs":
      return { routeName: "Logs" };
    default:
      return { routeName: "Overview" };
  }
};
