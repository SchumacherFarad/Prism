'use client';

import { motion, AnimatePresence } from 'framer-motion';
import { Trash2, Loader2 } from 'lucide-react';
import { Modal } from './Modal';
import { useState } from 'react';

interface DeleteConfirmProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => Promise<void>;
  itemName: string;
  itemType: 'fund' | 'crypto';
}

export function DeleteConfirm({ isOpen, onClose, onConfirm, itemName, itemType }: DeleteConfirmProps) {
  const [isDeleting, setIsDeleting] = useState(false);
  
  const handleConfirm = async () => {
    setIsDeleting(true);
    try {
      await onConfirm();
      onClose();
    } catch {
      // Error handling is done by the mutation
    } finally {
      setIsDeleting(false);
    }
  };
  
  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Delete Holding">
      <div className="space-y-4">
        <div className="flex items-center justify-center">
          <div className="w-16 h-16 rounded-full bg-red-500/10 flex items-center justify-center">
            <Trash2 size={32} className="text-red-400" />
          </div>
        </div>
        
        <div className="text-center">
          <p className="text-gray-300">
            Are you sure you want to delete
          </p>
          <p className="text-white font-semibold text-lg mt-1">
            {itemName}
          </p>
          <p className="text-gray-500 text-sm mt-2">
            This action cannot be undone. The {itemType === 'crypto' ? 'crypto' : 'fund'} will be removed from your portfolio.
          </p>
        </div>
        
        <div className="flex gap-3 pt-2">
          <button
            type="button"
            onClick={onClose}
            disabled={isDeleting}
            className="
              flex-1 px-4 py-3 rounded-lg
              bg-white/5 text-gray-300
              hover:bg-white/10
              disabled:opacity-50
              transition-colors
            "
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleConfirm}
            disabled={isDeleting}
            className="
              flex-1 px-4 py-3 rounded-lg
              bg-red-600 text-white font-medium
              hover:bg-red-500
              disabled:opacity-50 disabled:cursor-not-allowed
              transition-colors
              flex items-center justify-center gap-2
            "
          >
            {isDeleting && <Loader2 size={18} className="animate-spin" />}
            Delete
          </button>
        </div>
      </div>
    </Modal>
  );
}
