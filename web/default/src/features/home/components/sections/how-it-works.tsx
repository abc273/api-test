import { Settings, Zap, BarChart3 } from 'lucide-react'
import { AnimateInView } from '@/components/animate-in-view'

export function HowItWorks() {
  const steps = [
    {
      num: '1',
      title: '创建 API Key',
      desc: '在控制台创建 API Key，确认你要接入的模型和工作流。',
      icon: <Settings className='size-6' strokeWidth={1.5} />,
    },
    {
      num: '2',
      title: '调用统一接口',
      desc: '对话走 OpenAI 兼容接口，视频、图片和资产走对应业务接口。',
      icon: <Zap className='size-6' strokeWidth={1.5} />,
    },
    {
      num: '3',
      title: '查询结果并管理资产',
      desc: '在控制台查询任务状态、下载输出文件，并管理真人或虚拟资产。',
      icon: <BarChart3 className='size-6' strokeWidth={1.5} />,
    },
  ]

  return (
    <section className='border-border/40 relative z-10 border-t px-6 py-24 md:py-32'>
      <div className='mx-auto max-w-6xl'>
        <AnimateInView className='mb-16 text-center md:mb-20'>
          <p className='text-muted-foreground mb-3 text-xs font-medium tracking-widest uppercase'>
            接入流程
          </p>
          <h2 className='text-2xl font-bold tracking-tight md:text-3xl'>
            三步完成接入
          </h2>
        </AnimateInView>

        <div className='relative grid gap-8 md:grid-cols-3 md:gap-12'>
          {/* Connecting line (desktop) */}
          <div
            aria-hidden
            className='from-border/0 via-border to-border/0 absolute top-12 right-[20%] left-[20%] hidden h-px bg-gradient-to-r md:block'
          />

          {steps.map((step, i) => (
            <AnimateInView
              key={step.num}
              delay={i * 150}
              animation='fade-up'
              className='relative flex flex-col items-center text-center'
            >
              <div className='relative mb-6'>
                <div className='text-muted-foreground border-border/50 bg-muted/30 flex size-16 items-center justify-center rounded-2xl border transition-colors'>
                  {step.icon}
                </div>
                <div className='bg-foreground text-background absolute -top-2 -right-2 flex size-6 items-center justify-center rounded-full text-xs font-bold'>
                  {step.num}
                </div>
              </div>
              <h3 className='mb-2 text-base font-semibold'>{step.title}</h3>
              <p className='text-muted-foreground max-w-[240px] text-sm leading-relaxed'>
                {step.desc}
              </p>
            </AnimateInView>
          ))}
        </div>
      </div>
    </section>
  )
}
