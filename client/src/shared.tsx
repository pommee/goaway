export function NoContent({ text }: { text: string }) {
  return (
    <div className="flex flex-col items-center justify-center">
      <div className="relative bg-muted rounded-full p-3 mb-2">
        <img
          src="/gray-icon.png"
          alt="project-mascot"
          className="w-16 h-16 object-contain"
        />
      </div>
      <p className="text-sm text-muted-foreground">{text}</p>
    </div>
  );
}
