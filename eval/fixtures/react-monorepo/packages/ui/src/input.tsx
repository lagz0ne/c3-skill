interface InputProps {
  placeholder?: string;
  type?: "text" | "email" | "password";
  value?: string;
  onChange?: (value: string) => void;
}

export function Input({ placeholder, type = "text", value, onChange }: InputProps) {
  return (
    <input
      type={type}
      placeholder={placeholder}
      value={value}
      onChange={(e) => onChange?.(e.target.value)}
      className="input"
    />
  );
}
