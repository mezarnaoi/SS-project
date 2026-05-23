import React, { useState } from 'react';
import fallbackImage from '../../assets/photo-fallback.svg';
import { apiFetch } from '../../utils/api';

interface PhotoCardProps {
  photoId: string;
  imageUrl: string;
  altText?: string;
  extractedText?: string;
  isAdmin?: boolean;
  onDelete?: (photoId: string) => Promise<void> | void;
  needsReview?: boolean;
  reviewReason?: string;
  ocrConfidence?: number;
  reviewedBy?: string;
  token?: string;
  onReviewed?: (photoId: string) => void;
}

const PhotoCard: React.FC<PhotoCardProps> = ({
  photoId,
  imageUrl,
  altText = 'Photo',
  extractedText = '',
  isAdmin = false,
  onDelete,
  needsReview = false,
  reviewReason = '',
  ocrConfidence = 0,
  reviewedBy = '',
  token = '',
  onReviewed,
}) => {
  const [isZoomed, setIsZoomed] = useState(false);
  const [imageError, setImageError] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [isReviewing, setIsReviewing] = useState(false);
  const [reviewed, setReviewed] = useState(!!reviewedBy);

  const handleImageError = () => setImageError(true);
  const toggleZoom = () => setIsZoomed(!isZoomed);

  const handleModalClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) setIsZoomed(false);
  };

  const handleDeleteClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    setShowDeleteConfirm(true);
  };

  const handleConfirmDelete = async () => {
    setIsDeleting(true);
    if (onDelete) await onDelete(photoId);
    setShowDeleteConfirm(false);
    setIsDeleting(false);
  };

  const handleMarkReviewed = async (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsReviewing(true);
    try {
      const response = await apiFetch(`/photos/review/${photoId}`, {
        method: 'PATCH',
        headers: { Authorization: `Bearer ${token}` },
      });
      if (response.ok) {
        setReviewed(true);
        if (onReviewed) onReviewed(photoId);
      }
    } catch (err) {
      console.error('Failed to mark as reviewed', err);
    } finally {
      setIsReviewing(false);
    }
  };

  // Helper pentru a formata raw text în JSON pretty-print
  const formatOCRText = (text: string) => {
    if (!text) return '';
    try {
      const parsedObject = JSON.parse(text);
      return JSON.stringify(parsedObject, null, 2);
    } catch (error) {
      // Fallback dacă textul nu este un JSON valid
      return text;
    }
  };

  const confidenceColor =
    ocrConfidence >= 95
      ? 'text-green-600'
      : ocrConfidence >= 75
      ? 'text-yellow-600'
      : 'text-red-600';

  return (
    <>
      <div className="bg-white rounded-lg shadow-md overflow-hidden transition-all hover:shadow-lg relative">

        {/* ── Needs-Review badge ────────────────────────────────────────────── */}
        {needsReview && !reviewed && (
          <div className="absolute top-0 left-0 right-0 z-10 bg-orange-500 text-white text-xs font-semibold px-2 py-1 flex items-center justify-between">
            <span>⚠ Needs Human Review</span>
            <span className={`font-mono ${confidenceColor} bg-white rounded px-1`}>
              {ocrConfidence.toFixed(1)}%
            </span>
          </div>
        )}
        {reviewed && (
          <div className="absolute top-0 left-0 right-0 z-10 bg-green-500 text-white text-xs font-semibold px-2 py-1">
            ✓ Reviewed
          </div>
        )}

        <div
          className={`relative h-48 cursor-pointer ${needsReview && !reviewed ? 'mt-6' : reviewed ? 'mt-6' : ''}`}
          onClick={toggleZoom}
        >
          <img
            src={imageError ? fallbackImage : imageUrl}
            alt={altText}
            onError={handleImageError}
            className="w-full h-full object-cover"
          />
          {isAdmin && (
            <button
              onClick={handleDeleteClick}
              className="absolute top-2 right-2 bg-red-500 hover:bg-red-600 text-white rounded-full p-2 shadow-lg transition-all duration-200 opacity-80 hover:opacity-100"
              title="Delete photo"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          )}
        </div>

        {/* OCR confidence score */}
        {ocrConfidence > 0 && (
          <div className="px-3 pt-2 flex items-center gap-1">
            <span className="text-xs text-gray-400">OCR confidence:</span>
            <span className={`text-xs font-semibold ${confidenceColor}`}>
              {ocrConfidence.toFixed(1)}%
            </span>
          </div>
        )}

        {extractedText && (
          <div className="p-3 border-t border-gray-100">
            <p className="text-sm text-gray-600 truncate">{extractedText}</p>
          </div>
        )}

        {/* Mark-as-reviewed button */}
        {needsReview && !reviewed && (
          <div className="px-3 pb-3">
            <button
              onClick={handleMarkReviewed}
              disabled={isReviewing}
              className="w-full py-1.5 text-xs font-semibold bg-orange-100 hover:bg-orange-200 text-orange-800 rounded-md transition-colors disabled:opacity-50"
            >
              {isReviewing ? 'Saving…' : 'Mark as Reviewed'}
            </button>
            {reviewReason && (
              <p className="mt-1 text-xs text-gray-400 leading-tight">{reviewReason}</p>
            )}
          </div>
        )}

        {/* Delete confirmation dialog */}
        {showDeleteConfirm && (
          <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center">
            <div className="bg-white rounded-lg p-4 m-4 shadow-xl">
              <p className="text-gray-800 mb-4">Delete this photo?</p>
              <div className="flex gap-2 justify-center">
                <button
                  onClick={() => setShowDeleteConfirm(false)}
                  className="px-4 py-2 bg-gray-300 hover:bg-gray-400 rounded-md transition-colors"
                  disabled={isDeleting}
                >
                  Cancel
                </button>
                <button
                  onClick={handleConfirmDelete}
                  className="px-4 py-2 bg-red-500 hover:bg-red-600 text-white rounded-md transition-colors"
                  disabled={isDeleting}
                >
                  {isDeleting ? 'Deleting…' : 'Delete'}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Zoom modal */}
      {isZoomed && (
        <div
          className="fixed inset-0 bg-black bg-opacity-60 backdrop-blur-sm flex items-center justify-center z-50 p-4"
          onClick={handleModalClick}
        >
          {/* Am adăugat flex flex-col pentru a face scrollbar-ul să funcționeze corect */}
          <div className="relative bg-white rounded-xl shadow-2xl w-full max-w-4xl max-h-[90vh] flex flex-col overflow-hidden">
            <div className="absolute top-0 right-0 left-0 bg-gradient-to-b from-black/50 to-transparent h-20 z-10 flex justify-between items-start p-4 pointer-events-none">
              <div className="text-white text-lg font-medium truncate pr-10">{altText}</div>
              <button
                className="bg-white/20 hover:bg-white/40 text-white rounded-full p-2 backdrop-blur-sm transition-all duration-200 pointer-events-auto"
                onClick={toggleZoom}
              >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            {/* Review badge inside modal */}
            {needsReview && !reviewed && (
              <div className="absolute top-14 left-4 z-20 bg-orange-500 text-white text-xs font-semibold px-2 py-1 rounded">
                ⚠ Needs Review — OCR {ocrConfidence.toFixed(1)}%
              </div>
            )}

            {/* Containerul imaginii - înălțime fixată pentru a lăsa loc textului */}
            <div className="p-4 pt-20 flex-shrink-0 flex justify-center bg-gray-50 border-b border-gray-100">
              <img
                src={imageError ? fallbackImage : imageUrl}
                alt={altText}
                className="max-w-full max-h-[45vh] object-contain rounded-md"
              />
            </div>

            {/* Containerul de text scrollabil unde aplicăm formatarea JSON */}
            {extractedText && (
              <div className="bg-white p-6 overflow-y-auto flex-1">
                <h3 className="text-sm font-medium text-gray-500 mb-3">Extracted Data (JSON)</h3>
                <pre className="text-gray-800 text-sm font-mono whitespace-pre-wrap break-words bg-gray-50 p-4 rounded-md border border-gray-200 shadow-inner">
                  {formatOCRText(extractedText)}
                </pre>
              </div>
            )}

            {reviewReason && (
              <div className="bg-orange-50 px-6 py-4 border-t border-orange-100 flex-shrink-0">
                <p className="text-xs font-semibold text-orange-800">Review Reason</p>
                <p className="text-sm text-orange-700 mt-1">{reviewReason}</p>
              </div>
            )}
          </div>
        </div>
      )}
    </>
  );
};

export default PhotoCard;