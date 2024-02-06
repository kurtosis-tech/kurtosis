import { Button, Icon, Text } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { useMemo } from "react";
import { IoLogoDocker } from "react-icons/io5";

function getUrlForImage(image: string): string | null {
  const [imageName] = image.split(":");
  const imageParts = imageName.split("/");
  if (imageParts.length === 1) {
    return `https://hub.docker.com/_/${imageParts[0]}`;
  }
  if (imageParts.length === 2) {
    return `https://hub.docker.com/r/${imageParts[0]}/${imageParts[1]}`;
  }
  // Currently no other registries supported
  return null;
}

type ImageButtonProps = {
  image: string;
};

export const ImageButton = ({ image }: ImageButtonProps) => {
  const url = useMemo(() => getUrlForImage(image), [image]);

  if (!isDefined(url)) {
    return <Text fontSize={"xs"}>{image}</Text>;
  }

  return (
    <a href={url} target="_blank" rel="noopener noreferrer">
      <Button leftIcon={<Icon as={IoLogoDocker} color={"gray.400"} />} variant={"ghost"} size={"xs"}>
        {image}
      </Button>
    </a>
  );
};
