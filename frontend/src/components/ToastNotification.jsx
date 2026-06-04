// components/ToastNotification.jsx

export default function ToastNotification({
  toasts,
}) {
  return (
    <div className="toast-wrapper">

      {toasts.map((toast) => (

        <div
          key={toast.id}
          className="toast-item"
        >

          🔥 {toast.user} just bought Area {toast.area} ticket

        </div>

      ))}

    </div>
  );
}