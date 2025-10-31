import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import { GetRequest, PostRequest } from "@/util";
import {
  EyeClosedIcon,
  EyeIcon,
  LockIcon,
  SpinnerIcon,
  UserCircleIcon
} from "@phosphor-icons/react";
import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";

import { Metrics } from "@/components/server-statistics";
import { type ISourceOptions } from "@tsparticles/engine";
import { loadStarsPreset } from "@tsparticles/preset-stars";
import Particles, { initParticlesEngine } from "@tsparticles/react";

function FloatingTitle({ quote }: { quote: string }) {
  return (
    <div className="relative animate-float">
      <p className="font-bold animate-glow text-blue-300 text-6xl">GoAway</p>
      <p className="text-white">{quote}</p>
    </div>
  );
}

const FloatingParticles = () => {
  const [, setInit] = useState(false);

  useEffect(() => {
    initParticlesEngine(async (engine) => {
      await loadStarsPreset(engine);
    }).then(() => {
      setInit(true);
    });
  }, []);

  const options: ISourceOptions = {
    preset: "stars"
  };

  return <Particles id="tsparticles" options={options} />;
};

interface LoginProps extends React.ComponentPropsWithoutRef<"div"> {
  quote: string;
}

export default function Login({ className, quote, ...props }: LoginProps) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [responseData, setResponseData] = useState<Metrics>();
  const passwordRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    if (!username || !password) {
      toast.error("Please fill in both fields.");
      setIsLoading(false);
      return;
    }

    try {
      const [statusCode, response] = await PostRequest(
        "login",
        {
          username,
          password
        },
        true,
        true
      );

      if (statusCode === 200) {
        if (rememberMe) {
          localStorage.setItem("loginUsername", username);
        }

        navigate("/");
      } else if (statusCode === 429) {
        toast.warning("Rate limit exceeded", {
          description: `Retry again in ${response.retryAfterSeconds} seconds`
        });
        return;
      } else {
        toast.warning("Login failed", { description: response.error });
      }
    } catch (error) {
      console.error("Login error:", error);
      toast.error("Failed to login.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    async function fetchData() {
      try {
        const [, data] = await GetRequest("server");
        setResponseData(data);
      } catch {
        return;
      }
    }

    const rememberedLoginUsername = localStorage.getItem("loginUsername");
    if (rememberedLoginUsername) {
      setUsername(rememberedLoginUsername);
      setRememberMe(true);

      setTimeout(() => {
        passwordRef.current?.focus();
      }, 0);
    }

    fetchData();
  }, []);

  const togglePasswordVisibility = () => setShowPassword(!showPassword);

  return (
    <div className="flex min-h-screen w-full items-center justify-center p-4 overflow-hidden">
      <FloatingParticles />
      <div className="w-full max-w-md text-center">
        <FloatingTitle quote={quote} />

        <div className={cn("flex flex-col", className)} {...props}>
          <Card className="z-10 mt-10 border shadow-xl backdrop-blur-lg transition-all duration-300 hover:shadow-glow animate-card-appear">
            <CardContent className="pt-6">
              <form onSubmit={handleSubmit} className="space-y-6">
                <div className="flex flex-col gap-5">
                  <div className="space-y-2">
                    <Label
                      htmlFor="username"
                      className="text-sm font-medium text-muted-foreground"
                    >
                      Username
                    </Label>
                    <div className="relative group">
                      <UserCircleIcon className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                      <Input
                        id="username"
                        type="text"
                        placeholder="Enter your username"
                        required
                        autoFocus
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        className="pl-10"
                      />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label
                        htmlFor="password"
                        className="text-sm font-medium text-muted-foreground"
                      >
                        Password
                      </Label>
                    </div>
                    <div className="relative group">
                      <LockIcon className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                      <Input
                        id="password"
                        ref={passwordRef}
                        type={showPassword ? "text" : "password"}
                        placeholder="Enter your password"
                        required
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="pl-10 pr-10"
                      />
                      <button
                        type="button"
                        onClick={togglePasswordVisibility}
                        className="absolute right-3 top-3 text-muted-foreground focus:outline-none cursor-pointer"
                      >
                        {showPassword ? (
                          <EyeClosedIcon className="h-4 w-4" />
                        ) : (
                          <EyeIcon className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                  </div>

                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="remember"
                      checked={rememberMe}
                      onCheckedChange={(checked) => {
                        const isChecked = checked === true;
                        setRememberMe(isChecked);
                        if (!isChecked) {
                          localStorage.removeItem("loginUsername");
                        }
                      }}
                    />

                    <Label
                      htmlFor="remember"
                      className="text-sm font-medium leading-none text-muted-foreground cursor-pointer"
                    >
                      Remember me
                    </Label>
                  </div>

                  <Button
                    type="submit"
                    className="w-full bg-green-900 hover:bg-green-700 transition-all duration-300 hover:shadow-md hover:shadow-green-900/30 hover:translate-y-px animate-button-pulse focus:ring-2 focus:ring-green-700/50 disabled:opacity-70 text-white"
                    disabled={isLoading}
                  >
                    {isLoading ? (
                      <span className="flex items-center justify-center">
                        <SpinnerIcon className="animate-spin" />
                        Signing in...
                      </span>
                    ) : (
                      "Sign In"
                    )}
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>

          <div className="mt-6 text-muted-foreground text-sm z-10">
            <p>
              Version {responseData?.version} - Last updated{" "}
              {responseData?.date ? (
                new Date(responseData.date).toLocaleString("en-US", {
                  year: "numeric",
                  month: "short",
                  day: "numeric"
                })
              ) : (
                <span className="text-red-800">unavailable</span>
              )}
            </p>
            <p className="mt-1">
              <a
                href="https://github.com/pommee/goaway"
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-400 hover:text-blue-300 hover:underline transition-all duration-200"
              >
                View on GitHub
              </a>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
