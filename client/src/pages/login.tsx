import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import { GetRequest, PostRequest } from "@/util";
import {
  EyeIcon,
  EyeClosedIcon,
  LockIcon,
  SpinnerIcon,
  UserCircleIcon
} from "@phosphor-icons/react";
import { memo, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";

import { Metrics } from "@/components/server-statistics";
import { type ISourceOptions } from "@tsparticles/engine";
import { loadStarsPreset } from "@tsparticles/preset-stars";
import Particles, { initParticlesEngine } from "@tsparticles/react";

const FloatingTitle = () => {
  return (
    <h1 className="text-6xl font-bold text-indigo-300 relative inline-block animate-float">
      <span className="animate-glow">GoAway</span>
    </h1>
  );
};

const MemoizedParticles = () => {
  const [, setInit] = useState(false);

  useEffect(() => {
    initParticlesEngine(async (engine) => {
      await loadStarsPreset(engine);
    }).then(() => {
      setInit(true);
    });
  }, []);

  const options: ISourceOptions = useMemo(
    () => ({
      preset: "stars"
    }),
    []
  );

  return <Particles id="tsparticles" options={options} />;
};
const MP = memo(MemoizedParticles);

export function Login({
  className,
  ...props
}: React.ComponentPropsWithoutRef<"div">) {
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
    <div className="flex min-h-screen w-full items-center bg-zinc-900 justify-center p-4 overflow-hidden">
      <MP />
      <div className="w-full max-w-md text-center">
        <FloatingTitle />

        <div className={cn("flex flex-col", className)} {...props}>
          <Card className="z-10 mt-10 border border-zinc-800 shadow-xl bg-card-gradient backdrop-blur-lg transition-all duration-300 hover:shadow-glow animate-card-appear">
            <CardContent className="pt-6">
              <form onSubmit={handleSubmit} className="space-y-6">
                <div className="flex flex-col gap-5">
                  <div className="space-y-2">
                    <Label
                      htmlFor="username"
                      className="text-sm font-medium text-zinc-300"
                    >
                      Username
                    </Label>
                    <div className="relative group">
                      <UserCircleIcon className="absolute left-3 top-3 h-4 w-4 text-zinc-400 group-hover:text-indigo-300 transition-colors duration-300" />
                      <Input
                        id="username"
                        type="text"
                        placeholder="Enter your username"
                        required
                        autoFocus
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        className="pl-10 bg-zinc-900/70 border-zinc-700 text-zinc-200 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/50 transition-all duration-300 animate-input-appear placeholder:text-zinc-500"
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
                    <div className="relative group">
                      <LockIcon className="absolute left-3 top-3 h-4 w-4 text-zinc-400 group-hover:text-indigo-300 transition-colors duration-300" />
                      <Input
                        id="password"
                        ref={passwordRef}
                        type={showPassword ? "text" : "password"}
                        placeholder="Enter your password"
                        required
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="pl-10 pr-10 bg-zinc-900/70 border-zinc-700 text-zinc-200 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/50 transition-all duration-300 animate-input-appear placeholder:text-zinc-500"
                      />
                      <button
                        type="button"
                        onClick={togglePasswordVisibility}
                        className="absolute right-3 top-3 text-zinc-400 hover:text-indigo-300 transition-colors duration-200 focus:outline-none focus:text-indigo-400"
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
                      className="text-sm font-medium leading-none text-white cursor-pointer"
                    >
                      Remember me
                    </Label>
                  </div>

                  <Button
                    type="submit"
                    className="w-full bg-green-900 text-white hover:bg-green-700 transition-all duration-300 hover:shadow-md hover:shadow-green-900/30 hover:translate-y-px animate-button-pulse focus:ring-2 focus:ring-green-700/50 disabled:opacity-70"
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

          <div className="mt-6 text-zinc-500 text-sm z-10">
            <p>
              Version {responseData?.version} - Last updated{" "}
              {new Date(responseData?.date).toLocaleString("en-US", {
                year: "numeric",
                month: "short",
                day: "numeric"
              })}
            </p>
            <p className="mt-1">
              <a
                href="https://github.com/pommee/goaway"
                target="_blank"
                rel="noopener noreferrer"
                className="text-indigo-400 hover:text-indigo-300 hover:underline transition-all duration-200"
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
