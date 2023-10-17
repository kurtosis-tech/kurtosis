import ReactMarkdown from "react-markdown";
import remarkGfm from 'remark-gfm'
import {Text} from "@chakra-ui/react";

const PackageCatalogDescription = ({content}) => {
  return (
    <div className="markdown bg-white prose m-5" style={{height:"100%"}}>
      <ReactMarkdown children={content} remarkPlugins={[remarkGfm]}/>
    </div>
  );
};

export default PackageCatalogDescription;
