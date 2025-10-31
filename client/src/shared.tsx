export function NoContent({ text }: { text: string }) {
  return (
    <div className="flex flex-col items-center justify-center">
      <img
        src="/gray-icon.png"
        alt="project-mascot"
        className="w-16 h-16 mb-1 object-contain"
      />
      <p className="text-sm text-muted-foreground">{text}</p>
    </div>
  );
}
