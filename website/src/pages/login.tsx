import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { toast } from "sonner";
import { PostRequest } from "@/util";
import { useNavigate } from "react-router-dom";
import { Eye, EyeClosed, Lock, UserCircle } from "@phosphor-icons/react";
import { GenerateQuote } from "@/quotes";
import { TextAnimate } from "@/components/ui/text-animate";

export function Login({
  className,
  ...props
}: React.ComponentPropsWithoutRef<"div">) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [quote, setQuote] = useState(() => GenerateQuote());
  const navigate = useNavigate();

  useEffect(() => {
    const interval = setInterval(() => {
      setQuote(GenerateQuote());
    }, 4000);

    return () => clearInterval(interval);
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    if (!username || !password) {
      toast.error("Please fill in both fields.");
      setIsLoading(false);
      return;
    }

    try {
      const [statusCode] = await PostRequest("login", {
        username,
        password
      });

      if (statusCode === 200) {
        toast.success("Login successful!");
        navigate("/");
      } else {
        toast.error("Invalid username or password.");
      }
    } catch (error) {
      console.error("Login error:", error);
      toast.error("Failed to login.");
    } finally {
      setIsLoading(false);
    }
  };

  const togglePasswordVisibility = () => setShowPassword(!showPassword);

  return (
    <div className="flex min-h-screen w-full items-center bg-zinc-900 justify-center p-4">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold text-indigo-300">GoAway</h1>
          <TextAnimate
            className="truncate text-xs text-zinc-500"
            animation="blurInUp"
            by="character"
            once
          >
            {quote}
          </TextAnimate>
        </div>

        <div className={cn("flex flex-col", className)} {...props}>
          <Card className="border border-zinc-800 shadow-xl bg-zinc-950">
            <CardHeader className="space-y-1 pb-4">
              <CardTitle className="text-2xl font-bold text-center text-white">
                Login
              </CardTitle>
              <CardDescription className="text-center text-zinc-400">
                Enter your credentials to continue
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit}>
                <div className="flex flex-col gap-4">
                  <div className="space-y-2">
                    <Label
                      htmlFor="username"
                      className="text-sm font-medium text-zinc-300"
                    >
                      Username
                    </Label>
                    <div className="relative">
                      <UserCircle className="absolute left-3 top-3 h-4 w-4 text-zinc-400" />
                      <Input
                        id="username"
                        type="text"
                        placeholder="Enter your username"
                        required
                        autoFocus
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        className="pl-10 bg-zinc-900 border-zinc-800 text-zinc-200 focus:border-indigo-500 focus:ring-indigo-500"
                      />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label
                        htmlFor="password"
                        className="text-sm font-medium text-zinc-300"
                      >
                        Password
                      </Label>
                    </div>
                    <div className="relative">
                      <Lock className="absolute left-3 top-3 h-4 w-4 text-zinc-400" />
                      <Input
                        id="password"
                        type={showPassword ? "text" : "password"}
                        placeholder="Enter your password"
                        required
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="pl-10 pr-10 bg-zinc-900 border-zinc-800 text-zinc-200 focus:border-indigo-500 focus:ring-indigo-500"
                      />
                      <button
                        type="button"
                        onClick={togglePasswordVisibility}
                        className="absolute right-3 top-3 text-zinc-400 hover:text-indigo-300"
                      >
                        {showPassword ? (
                          <EyeClosed className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                  </div>

                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="remember"
                      checked={rememberMe}
                      onCheckedChange={(checked) =>
                        setRememberMe(checked as boolean)
                      }
                      className="border-zinc-700 data-[state=checked]:bg-indigo-600 data-[state=checked]:border-indigo-600"
                    />
                    <Label
                      htmlFor="remember"
                      className="text-sm font-medium leading-none text-zinc-300 peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                    >
                      Remember me
                    </Label>
                  </div>

                  <Button
                    type="submit"
                    className="w-full bg-green-900 text-white hover:bg-green-700 transition-colors"
                    disabled={isLoading}
                  >
                    {isLoading ? "Signing in..." : "Sign In"}
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
