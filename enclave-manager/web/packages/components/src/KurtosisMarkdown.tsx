import { Code, Divider, Heading, Image, Link, Table, Tbody, Td, Text, Th, Thead, Tr } from "@chakra-ui/react";
import { DetailedHTMLProps, HTMLAttributes } from "react";
import Markdown, { Components } from "react-markdown";

const heading =
  (level: 1 | 2 | 3 | 4 | 5 | 6) =>
  ({ children }: DetailedHTMLProps<HTMLAttributes<HTMLHeadingElement>, HTMLHeadingElement>) => {
    const sizes = ["xl", "lg", "md", "sm", "xs", "xs"];
    return (
      <Heading my={4} as={`h${level}`} size={sizes[`${level - 1}`]}>
        {children}
      </Heading>
    );
  };

const componentStrategy: Components = {
  h1: heading(1),
  h2: heading(2),
  h3: heading(3),
  h4: heading(4),
  h5: heading(5),
  h6: heading(6),
  p: (props) => {
    const { children } = props;
    return <Text mb={2}>{children}</Text>;
  },
  em: (props) => {
    const { children } = props;
    return <Text as="em">{children}</Text>;
  },
  blockquote: (props) => {
    const { children } = props;
    return (
      <Code as="blockquote" p={2}>
        {children}
      </Code>
    );
  },
  code: ({ children }) => {
    return <Code children={children} />;
  },
  del: (props) => {
    const { children } = props;
    return <Text as="del">{children}</Text>;
  },
  hr: (props) => {
    return <Divider />;
  },
  a: Link,
  img: (props) => <Image src={props.src} />,
  text: (props) => {
    const { children } = props;
    return <Text as="span">{children}</Text>;
  },
  pre: (props) => {
    const { children } = props;
    return (
      <Text margin={1} as={"pre"}>
        {children}
      </Text>
    );
  },
  table: Table,
  thead: Thead,
  tbody: Tbody,
  tr: (props) => <Tr>{props.children}</Tr>,
  td: (props) => <Td>{props.children}</Td>,
  th: (props) => <Th>{props.children}</Th>,
};

type KurtosisMarkdownProps = {
  children?: string;
};

export const KurtosisMarkdown = ({ children }: KurtosisMarkdownProps) => {
  return (
    <Markdown components={componentStrategy} skipHtml>
      {children}
    </Markdown>
  );
};
