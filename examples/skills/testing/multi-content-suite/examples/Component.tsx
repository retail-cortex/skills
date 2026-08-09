import React from 'react';

interface AssetViewerProps {
  title: string;
  mimeType: string;
  byteSize: number;
}

export const AssetViewer: React.FC<AssetViewerProps> = ({ title, mimeType, byteSize }) => {
  return (
    <div className="p-4 border rounded-lg shadow-sm">
      <h3 className="text-lg font-bold">{title}</h3>
      <span className="text-sm text-gray-500">{mimeType} ({byteSize} bytes)</span>
    </div>
  );
};
