;; test001.lisp
;; Time-stamp: <2017-10-02 21:04:25 krylon>
;;
;; I am going to try to implement a few primitive Lisp operations in pure
;; Lisp to see how that works out.
;; I will probably replace with them with special forms implemented in Go,
;; if only for performance reasons, but I want to try, for a moment,
;; how far I could push this implementing-Lisp-in-Lisp-thing.

(defun nil? (x)
  (eq x nil))

(defun map (fn lst)
  (if (nil? lst) nil
    (cons (fn (car lst))
	  (map fn (cdr lst)))))
