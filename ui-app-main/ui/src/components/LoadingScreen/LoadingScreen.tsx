import * as React from "react";
import { LoaderCircle } from "lucide-react";

type LoadingProps = {
  title?: string[];
  description?: string[];
};

export default function LoadingScreen({
  title = [""],
  description = [""],
}: LoadingProps) {
  const renderText = (text: string[]) => {
    return text.map((line, index) => (
      <React.Fragment key={index}>
        {line}
        {index < title.length - 1 && <br />}
      </React.Fragment>
    ));
  };

  return (
    <div className="min-h-screen bg-white flex items-center justify-center">
      <div className="w-full max-w-lg mx-auto px-4">
        <div className="flex flex-col items-center text-center space-y-6">
          <div className="w-16 h-16 flex items-center justify-center">
            <LoaderCircle className="w-12 h-12 text-[#3b5c3f] animate-spin" />
          </div>
          <div className="space-y-4">
            <h1 className="text-3xl font-bold">{renderText(title)}</h1>
            <p className="text-xl text-gray-500"> {renderText(description)}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
